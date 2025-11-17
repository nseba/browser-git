package protocol

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NegotiationRequest represents a client's want/have negotiation request
type NegotiationRequest struct {
	Wants        []string          // Commit hashes the client wants
	Haves        []string          // Commit hashes the client already has
	Capabilities []string          // Capabilities to request
	Deepen       int               // Depth for shallow clone (0 for full clone)
	Filters      map[string]string // Object filters (for partial clones)
	Done         bool              // Whether negotiation is complete
}

// NegotiationResponse represents the server's response to negotiation
type NegotiationResponse struct {
	ACKs      []ACK  // Acknowledgments from server
	NAK       bool   // Whether server sent NAK (negative acknowledgment)
	Packfile  []byte // Packfile data (if negotiation complete)
	SideBand  bool   // Whether response uses side-band protocol
	ErrorMsg  string // Error message if any
}

// ACKStatus represents the status of an ACK
type ACKStatus string

const (
	// ACKCommon indicates the commit is common to both client and server
	ACKCommon ACKStatus = "common"
	// ACKReady indicates the server is ready to send packfile
	ACKReady ACKStatus = "ready"
	// ACKContinue indicates negotiation should continue
	ACKContinue ACKStatus = "continue"
	// ACKSingle is a simple ACK without status
	ACKSingle ACKStatus = ""
)

// ACK represents a server acknowledgment
type ACK struct {
	Hash   string    // Commit hash being acknowledged
	Status ACKStatus // Status of the acknowledgment
}

// UploadPackClient handles the upload-pack protocol (fetch/clone)
type UploadPackClient struct {
	client  *Client
	repoURL string
}

// NewUploadPackClient creates a new upload-pack client
func NewUploadPackClient(client *Client, repoURL string) *UploadPackClient {
	return &UploadPackClient{
		client:  client,
		repoURL: repoURL,
	}
}

// Negotiate performs the want/have negotiation with the server
func (u *UploadPackClient) Negotiate(req *NegotiationRequest) (*NegotiationResponse, error) {
	// Build the upload-pack URL
	uploadPackURL, err := buildUploadPackURL(u.repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Encode the request body
	requestBody, err := encodeNegotiationRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest("POST", uploadPackURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", u.client.userAgent)
	httpReq.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	httpReq.Header.Set("Accept", "application/x-git-upload-pack-result")
	httpReq.Header.Set("Git-Protocol", "version=2")

	// Apply authentication
	if err := u.client.authProvider.ApplyAuth(httpReq); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Make the request
	resp, err := u.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("negotiation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("negotiation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	negotiationResp, err := parseNegotiationResponse(resp.Body, req.Done, hasSideBandCapability(req.Capabilities))
	if err != nil {
		return nil, fmt.Errorf("failed to parse negotiation response: %w", err)
	}

	return negotiationResp, nil
}

// FetchPackfile performs a complete negotiation and fetches the packfile
func (u *UploadPackClient) FetchPackfile(wants []string, haves []string, capabilities []string) ([]byte, error) {
	// Build negotiation request
	req := &NegotiationRequest{
		Wants:        wants,
		Haves:        haves,
		Capabilities: capabilities,
		Done:         true, // Complete negotiation in one round
	}

	// Perform negotiation
	resp, err := u.Negotiate(req)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.ErrorMsg != "" {
		return nil, fmt.Errorf("server error: %s", resp.ErrorMsg)
	}

	// Return packfile data
	return resp.Packfile, nil
}

// buildUploadPackURL constructs the upload-pack service URL
func buildUploadPackURL(repoURL string) (string, error) {
	// Parse and normalize the URL
	if !strings.HasSuffix(repoURL, "/") {
		if !strings.HasSuffix(repoURL, ".git") {
			repoURL += ".git"
		}
		repoURL += "/"
	}
	return repoURL + "git-upload-pack", nil
}

// encodeNegotiationRequest encodes a negotiation request to pkt-line format
func encodeNegotiationRequest(req *NegotiationRequest) ([]byte, error) {
	var buf bytes.Buffer
	writer := NewPktLineWriter(&buf)

	// Write want lines
	for i, want := range req.Wants {
		var line string
		if i == 0 && len(req.Capabilities) > 0 {
			// First want line includes capabilities
			capsStr := strings.Join(req.Capabilities, " ")
			line = fmt.Sprintf("want %s %s\n", want, capsStr)
		} else {
			line = fmt.Sprintf("want %s\n", want)
		}
		if err := writer.WriteString(line); err != nil {
			return nil, err
		}
	}

	// Handle deepen for shallow clones
	if req.Deepen > 0 {
		line := fmt.Sprintf("deepen %d\n", req.Deepen)
		if err := writer.WriteString(line); err != nil {
			return nil, err
		}
	}

	// Write flush after wants
	if err := writer.WriteFlush(); err != nil {
		return nil, err
	}

	// Write have lines if provided
	if len(req.Haves) > 0 {
		for _, have := range req.Haves {
			line := fmt.Sprintf("have %s\n", have)
			if err := writer.WriteString(line); err != nil {
				return nil, err
			}
		}
	}

	// Write done if negotiation is complete
	if req.Done {
		if err := writer.WriteString("done\n"); err != nil {
			return nil, err
		}
	} else {
		// Otherwise write flush to wait for server response
		if err := writer.WriteFlush(); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// parseNegotiationResponse parses the server's negotiation response
func parseNegotiationResponse(body io.Reader, done bool, sideBand bool) (*NegotiationResponse, error) {
	reader := NewPktLineReader(body)
	response := &NegotiationResponse{
		ACKs:     []ACK{},
		SideBand: sideBand,
	}

	// If side-band is enabled, we need to demultiplex the stream
	if sideBand {
		return parseSideBandResponse(reader, done)
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

		// Flush packet signals end of ACKs
		if line == nil {
			break
		}

		lineStr := string(line)
		lineStr = strings.TrimSuffix(lineStr, "\n")

		// Parse ACK lines
		if strings.HasPrefix(lineStr, "ACK ") {
			ack, err := parseACKLine(lineStr)
			if err != nil {
				return nil, err
			}
			response.ACKs = append(response.ACKs, ack)
		} else if lineStr == "NAK" {
			response.NAK = true
		} else if strings.HasPrefix(lineStr, "ERR ") {
			response.ErrorMsg = strings.TrimPrefix(lineStr, "ERR ")
			return response, nil
		}
	}

	// If negotiation is done, read the packfile data
	if done {
		packfile, err := io.ReadAll(reader.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read packfile: %w", err)
		}
		response.Packfile = packfile
	}

	return response, nil
}

// parseSideBandResponse parses a side-band multiplexed response
func parseSideBandResponse(reader *PktLineReader, done bool) (*NegotiationResponse, error) {
	response := &NegotiationResponse{
		ACKs:     []ACK{},
		SideBand: true,
	}

	var packfileBuf bytes.Buffer

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
		case 1: // Packfile data
			packfileBuf.Write(data)
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

	if done && packfileBuf.Len() > 0 {
		response.Packfile = packfileBuf.Bytes()
	}

	return response, nil
}

// parseACKLine parses an ACK line
// Format: "ACK <hash> [status]"
func parseACKLine(line string) (ACK, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ACK{}, fmt.Errorf("invalid ACK line: %s", line)
	}

	ack := ACK{
		Hash: parts[1],
	}

	// Parse optional status
	if len(parts) >= 3 {
		status := parts[2]
		switch status {
		case "common":
			ack.Status = ACKCommon
		case "ready":
			ack.Status = ACKReady
		case "continue":
			ack.Status = ACKContinue
		default:
			return ACK{}, fmt.Errorf("unknown ACK status: %s", status)
		}
	}

	return ack, nil
}

// hasSideBandCapability checks if side-band capability is requested
func hasSideBandCapability(capabilities []string) bool {
	for _, cap := range capabilities {
		if cap == "side-band" || cap == "side-band-64k" {
			return true
		}
	}
	return false
}

// BuildCapabilities builds a list of default capabilities for negotiation
func BuildCapabilities() []string {
	return []string{
		"multi_ack_detailed", // Detailed ACK responses
		"side-band-64k",      // Multiplexed output with 64k chunks
		"thin-pack",          // Allow thin packs
		"ofs-delta",          // Use offset deltas
		"agent=browser-git/0.1.0",
	}
}
