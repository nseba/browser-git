package protocol

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PushRequest represents a client's push request
type PushRequest struct {
	Updates      []RefUpdate       // Reference updates to push
	Capabilities []string          // Capabilities to request
	Packfile     []byte            // Packfile data with objects to send
	Force        bool              // Whether to allow non-fast-forward updates
	ReportStatus bool              // Whether to request status report
}

// RefUpdate represents a reference update for push
type RefUpdate struct {
	Name    string // Reference name (e.g., "refs/heads/main")
	OldHash string // Current hash on remote (40 zeros for new refs)
	NewHash string // New hash to push (40 zeros for delete)
}

// PushResponse represents the server's response to a push
type PushResponse struct {
	UnpackStatus string           // Status of unpacking (ok or error message)
	RefStatuses  []RefUpdateStatus // Status of each reference update
	SideBand     bool             // Whether response uses side-band protocol
	ErrorMsg     string           // Error message if any
}

// RefUpdateStatus represents the status of a reference update
type RefUpdateStatus struct {
	RefName string // Reference name
	Status  string // "ok" or error message
}

// ReceivePackClient handles the receive-pack protocol (push)
type ReceivePackClient struct {
	client  *Client
	repoURL string
}

// NewReceivePackClient creates a new receive-pack client
func NewReceivePackClient(client *Client, repoURL string) *ReceivePackClient {
	return &ReceivePackClient{
		client:  client,
		repoURL: repoURL,
	}
}

// Push performs a push operation to the remote repository
func (r *ReceivePackClient) Push(req *PushRequest) (*PushResponse, error) {
	// Build the receive-pack URL
	receivePackURL, err := buildReceivePackURL(r.repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Encode the request body
	requestBody, err := encodePushRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest("POST", receivePackURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", r.client.userAgent)
	httpReq.Header.Set("Content-Type", "application/x-git-receive-pack-request")
	httpReq.Header.Set("Accept", "application/x-git-receive-pack-result")
	httpReq.Header.Set("Git-Protocol", "version=2")

	// Apply authentication
	if err := r.client.authProvider.ApplyAuth(httpReq); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Make the request
	resp, err := r.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, WrapProtocolError(err, 0, r.repoURL)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("%s", string(body))
		return nil, WrapProtocolError(err, resp.StatusCode, r.repoURL)
	}

	// Parse the response
	pushResp, err := parsePushResponse(resp.Body, req.ReportStatus, hasSideBandCapability(req.Capabilities))
	if err != nil {
		return nil, fmt.Errorf("failed to parse push response: %w", err)
	}

	// Check for errors in response
	if pushResp.UnpackStatus != "ok" {
		return nil, fmt.Errorf("unpack failed: %s", pushResp.UnpackStatus)
	}

	// Check for reference update errors
	for _, refStatus := range pushResp.RefStatuses {
		if refStatus.Status != "ok" {
			return nil, fmt.Errorf("failed to update %s: %s", refStatus.RefName, refStatus.Status)
		}
	}

	return pushResp, nil
}

// buildReceivePackURL constructs the receive-pack service URL
func buildReceivePackURL(repoURL string) (string, error) {
	// Parse and normalize the URL
	if !strings.HasSuffix(repoURL, "/") {
		if !strings.HasSuffix(repoURL, ".git") {
			repoURL += ".git"
		}
		repoURL += "/"
	}
	return repoURL + "git-receive-pack", nil
}

// encodePushRequest encodes a push request to pkt-line format
func encodePushRequest(req *PushRequest) ([]byte, error) {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	// Write ref update lines
	for i, update := range req.Updates {
		var line string
		if i == 0 && len(req.Capabilities) > 0 {
			// First update line includes capabilities
			capsStr := strings.Join(req.Capabilities, " ")
			line = fmt.Sprintf("%s %s %s\x00%s\n", update.OldHash, update.NewHash, update.Name, capsStr)
		} else {
			line = fmt.Sprintf("%s %s %s\n", update.OldHash, update.NewHash, update.Name)
		}
		if err := writer.WriteString(line); err != nil {
			return nil, err
		}
	}

	// Write flush after updates
	if err := writer.WriteFlush(); err != nil {
		return nil, err
	}

	// Append packfile data if present
	if len(req.Packfile) > 0 {
		buf.Write(req.Packfile)
	}

	return buf.Bytes(), nil
}

// parsePushResponse parses the server's push response
func parsePushResponse(body io.Reader, reportStatus bool, sideBand bool) (*PushResponse, error) {
	// If no status report requested, consider it successful
	if !reportStatus {
		return &PushResponse{
			UnpackStatus: "ok",
			RefStatuses:  []RefUpdateStatus{},
		}, nil
	}

	reader := NewPktLineReader(body)
	response := &PushResponse{
		RefStatuses: []RefUpdateStatus{},
		SideBand:    sideBand,
	}

	// If side-band is enabled, we need to demultiplex the stream
	if sideBand {
		return parseSideBandPushResponse(reader)
	}

	// Standard response parsing
	for {
		line, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read line: %w", err)
		}

		// Flush packet signals end
		if line == nil {
			break
		}

		lineStr := string(line)
		lineStr = strings.TrimSuffix(lineStr, "\n")

		// Parse unpack status (first line)
		if strings.HasPrefix(lineStr, "unpack ") {
			response.UnpackStatus = strings.TrimPrefix(lineStr, "unpack ")
		} else if strings.HasPrefix(lineStr, "ok ") {
			// Successful reference update
			refName := strings.TrimPrefix(lineStr, "ok ")
			response.RefStatuses = append(response.RefStatuses, RefUpdateStatus{
				RefName: refName,
				Status:  "ok",
			})
		} else if strings.HasPrefix(lineStr, "ng ") {
			// Failed reference update
			// Format: "ng <refname> <error-message>"
			parts := strings.SplitN(strings.TrimPrefix(lineStr, "ng "), " ", 2)
			if len(parts) >= 2 {
				response.RefStatuses = append(response.RefStatuses, RefUpdateStatus{
					RefName: parts[0],
					Status:  parts[1],
				})
			}
		}
	}

	return response, nil
}

// parseSideBandPushResponse parses a side-band multiplexed push response
func parseSideBandPushResponse(reader *PktLineReader) (*PushResponse, error) {
	response := &PushResponse{
		RefStatuses: []RefUpdateStatus{},
		SideBand:    true,
	}

	var statusBuf bytes.Buffer

	for {
		line, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read side-band line: %w", err)
		}

		// Flush packet signals end
		if line == nil {
			break
		}

		if len(line) == 0 {
			continue
		}

		// First byte is the channel
		channel := line[0]
		data := line[1:]

		switch channel {
		case 1: // Status data
			statusBuf.Write(data)
		case 2: // Progress messages (stderr)
			// Progress messages can be logged or ignored
			// fmt.Fprintf(os.Stderr, "%s", string(data))
		case 3: // Error messages
			response.ErrorMsg = string(data)
			return response, nil
		default:
			return nil, fmt.Errorf("unknown side-band channel: %d", channel)
		}
	}

	// Parse status data
	if statusBuf.Len() > 0 {
		statusLines := strings.Split(statusBuf.String(), "\n")
		for _, line := range statusLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse unpack status
			if strings.HasPrefix(line, "unpack ") {
				response.UnpackStatus = strings.TrimPrefix(line, "unpack ")
			} else if strings.HasPrefix(line, "ok ") {
				// Successful reference update
				refName := strings.TrimPrefix(line, "ok ")
				response.RefStatuses = append(response.RefStatuses, RefUpdateStatus{
					RefName: refName,
					Status:  "ok",
				})
			} else if strings.HasPrefix(line, "ng ") {
				// Failed reference update
				parts := strings.SplitN(strings.TrimPrefix(line, "ng "), " ", 2)
				if len(parts) >= 2 {
					response.RefStatuses = append(response.RefStatuses, RefUpdateStatus{
						RefName: parts[0],
						Status:  parts[1],
					})
				}
			}
		}
	}

	return response, nil
}

// BuildPushCapabilities builds a list of default capabilities for push
func BuildPushCapabilities() []string {
	return []string{
		"report-status",           // Request status report
		"side-band-64k",           // Multiplexed output with 64k chunks
		"ofs-delta",               // Use offset deltas
		"agent=browser-git/0.1.0",
	}
}

// NewRefUpdate creates a new reference update
func NewRefUpdate(name, oldHash, newHash string) RefUpdate {
	return RefUpdate{
		Name:    name,
		OldHash: oldHash,
		NewHash: newHash,
	}
}

// NewRefUpdateForNew creates a reference update for a new branch
func NewRefUpdateForNew(name, newHash string) RefUpdate {
	return RefUpdate{
		Name:    name,
		OldHash: "0000000000000000000000000000000000000000", // 40 zeros for new ref
		NewHash: newHash,
	}
}

// NewRefUpdateForDelete creates a reference update to delete a branch
func NewRefUpdateForDelete(name, oldHash string) RefUpdate {
	return RefUpdate{
		Name:    name,
		OldHash: oldHash,
		NewHash: "0000000000000000000000000000000000000000", // 40 zeros for delete
	}
}
