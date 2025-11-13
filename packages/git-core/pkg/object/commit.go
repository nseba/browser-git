package object

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/nseba/browser-git/git-core/pkg/hash"
)

// Signature represents a Git signature (author or committer)
type Signature struct {
	Name  string
	Email string
	When  time.Time
}

// Commit represents a Git commit object
type Commit struct {
	Tree      hash.Hash
	Parents   []hash.Hash
	Author    Signature
	Committer Signature
	Message   string
	hash      hash.Hash
}

// NewCommit creates a new commit object
func NewCommit() *Commit {
	return &Commit{
		Parents: make([]hash.Hash, 0),
	}
}

// Type returns the object type (commit)
func (c *Commit) Type() Type {
	return CommitType
}

// Hash returns the hash of the commit
func (c *Commit) Hash() hash.Hash {
	return c.hash
}

// SetHash sets the hash of the commit
func (c *Commit) SetHash(h hash.Hash) {
	c.hash = h
}

// Size returns the size of the serialized commit content
func (c *Commit) Size() int64 {
	var buf bytes.Buffer
	_ = c.Serialize(&buf)
	return int64(buf.Len())
}

// AddParent adds a parent commit hash
func (c *Commit) AddParent(parent hash.Hash) {
	c.Parents = append(c.Parents, parent)
}

// IsRoot returns true if this is a root commit (no parents)
func (c *Commit) IsRoot() bool {
	return len(c.Parents) == 0
}

// IsMerge returns true if this is a merge commit (multiple parents)
func (c *Commit) IsMerge() bool {
	return len(c.Parents) > 1
}

// Serialize writes the commit content to a writer (without header)
func (c *Commit) Serialize(w io.Writer) error {
	// Write tree line
	if _, err := fmt.Fprintf(w, "tree %s\n", c.Tree.String()); err != nil {
		return err
	}

	// Write parent lines
	for _, parent := range c.Parents {
		if _, err := fmt.Fprintf(w, "parent %s\n", parent.String()); err != nil {
			return err
		}
	}

	// Write author line
	if _, err := fmt.Fprintf(w, "author %s\n", c.Author.Format()); err != nil {
		return err
	}

	// Write committer line
	if _, err := fmt.Fprintf(w, "committer %s\n", c.Committer.Format()); err != nil {
		return err
	}

	// Write empty line before message
	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	// Write message
	if _, err := w.Write([]byte(c.Message)); err != nil {
		return err
	}

	return nil
}

// SerializeWithHeader writes the complete commit (header + content) to a writer
func (c *Commit) SerializeWithHeader(w io.Writer) error {
	// Write header: "commit <size>\0"
	header := fmt.Sprintf("%s %d\x00", CommitType, c.Size())
	if _, err := w.Write([]byte(header)); err != nil {
		return err
	}

	// Write content
	return c.Serialize(w)
}

// Bytes returns the complete serialized commit (header + content)
func (c *Commit) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := c.SerializeWithHeader(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseCommit parses a commit from raw content data (without header)
func ParseCommit(data []byte) (*Commit, error) {
	commit := NewCommit()
	lines := strings.Split(string(data), "\n")

	// Parse header lines
	i := 0
	for i < len(lines) {
		line := lines[i]
		if line == "" {
			// Empty line separates headers from message
			i++
			break
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid commit line: %s", line)
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "tree":
			h, err := hash.ParseHash(value)
			if err != nil {
				return nil, fmt.Errorf("invalid tree hash: %w", err)
			}
			commit.Tree = h

		case "parent":
			h, err := hash.ParseHash(value)
			if err != nil {
				return nil, fmt.Errorf("invalid parent hash: %w", err)
			}
			commit.AddParent(h)

		case "author":
			sig, err := ParseSignature(value)
			if err != nil {
				return nil, fmt.Errorf("invalid author signature: %w", err)
			}
			commit.Author = sig

		case "committer":
			sig, err := ParseSignature(value)
			if err != nil {
				return nil, fmt.Errorf("invalid committer signature: %w", err)
			}
			commit.Committer = sig

		default:
			// Ignore unknown headers for forward compatibility
		}

		i++
	}

	// Parse message (remaining lines)
	if i < len(lines) {
		commit.Message = strings.Join(lines[i:], "\n")
	}

	return commit, nil
}

// ComputeHash computes and sets the hash of the commit using the given hasher
func (c *Commit) ComputeHash(hasher hash.Hasher) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	c.hash = hasher.Hash(data)
	return nil
}

// Format formats a signature as "Name <email> timestamp timezone"
func (s *Signature) Format() string {
	timestamp := s.When.Unix()
	_, offset := s.When.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60
	timezone := fmt.Sprintf("%+03d%02d", offsetHours, offsetMinutes)

	return fmt.Sprintf("%s <%s> %d %s", s.Name, s.Email, timestamp, timezone)
}

// ParseSignature parses a signature from a string
// Format: "Name <email> timestamp timezone"
func ParseSignature(s string) (Signature, error) {
	// Find email boundaries
	emailStart := strings.Index(s, "<")
	emailEnd := strings.Index(s, ">")

	if emailStart == -1 || emailEnd == -1 || emailEnd <= emailStart {
		return Signature{}, fmt.Errorf("invalid signature format: missing or malformed email")
	}

	// Parse name (before email)
	name := strings.TrimSpace(s[:emailStart])

	// Parse email
	email := s[emailStart+1 : emailEnd]

	// Parse timestamp and timezone (after email)
	remainder := strings.TrimSpace(s[emailEnd+1:])
	parts := strings.Fields(remainder)

	if len(parts) < 2 {
		return Signature{}, fmt.Errorf("invalid signature format: missing timestamp or timezone")
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Signature{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse timezone
	timezone := parts[1]
	if len(timezone) != 5 {
		return Signature{}, fmt.Errorf("invalid timezone format: %s", timezone)
	}

	sign := 1
	if timezone[0] == '-' {
		sign = -1
	}

	hours, err := strconv.Atoi(timezone[1:3])
	if err != nil {
		return Signature{}, fmt.Errorf("invalid timezone hours: %w", err)
	}

	minutes, err := strconv.Atoi(timezone[3:5])
	if err != nil {
		return Signature{}, fmt.Errorf("invalid timezone minutes: %w", err)
	}

	offset := sign * (hours*3600 + minutes*60)
	location := time.FixedZone("", offset)
	when := time.Unix(timestamp, 0).In(location)

	return Signature{
		Name:  name,
		Email: email,
		When:  when,
	}, nil
}

// String returns a string representation of the signature
func (s *Signature) String() string {
	return s.Format()
}
