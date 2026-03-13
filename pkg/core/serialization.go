package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	aierr "github.com/apshoemaker/ai-attr/pkg/errors"
)

const (
	AuthorshipLogVersion = "authorship/3.0.0"
	ToolVersion          = "ai-attr/0.1.0"
	HumanSentinel        = "human___________"
)

// AgentId identifies an AI agent.
type AgentId struct {
	Tool  string `json:"tool"`
	ID    string `json:"id"`
	Model string `json:"model"`
}

// PromptRecord stores prompt session details in the metadata section.
type PromptRecord struct {
	AgentID       AgentId `json:"agent_id"`
	HumanAuthor   string  `json:"human_author,omitempty"`
	TotalAdds     uint32  `json:"total_additions"`
	TotalDels     uint32  `json:"total_deletions"`
	AcceptedLines uint32  `json:"accepted_lines"`
}

// AuthorshipMetadata is the JSON metadata section below the divider.
type AuthorshipMetadata struct {
	SchemaVersion string                   `json:"schema_version"`
	ToolVersion   string                   `json:"tool_version"`
	BaseCommitSHA string                   `json:"base_commit_sha,omitempty"`
	Prompts       map[string]*PromptRecord `json:"prompts"`
}

func NewAuthorshipMetadata() AuthorshipMetadata {
	return AuthorshipMetadata{
		SchemaVersion: AuthorshipLogVersion,
		ToolVersion:   ToolVersion,
		Prompts:       make(map[string]*PromptRecord),
	}
}

// MarshalJSON ensures prompts keys are serialized in sorted order.
func (m AuthorshipMetadata) MarshalJSON() ([]byte, error) {
	type orderedMetadata struct {
		SchemaVersion string              `json:"schema_version"`
		ToolVersion   string              `json:"tool_version"`
		BaseCommitSHA string              `json:"base_commit_sha,omitempty"`
		Prompts       orderedPrompts      `json:"prompts"`
	}

	return json.Marshal(orderedMetadata{
		SchemaVersion: m.SchemaVersion,
		ToolVersion:   m.ToolVersion,
		BaseCommitSHA: m.BaseCommitSHA,
		Prompts:       orderedPrompts(m.Prompts),
	})
}

// orderedPrompts serializes map keys in sorted order.
type orderedPrompts map[string]*PromptRecord

func (op orderedPrompts) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(op))
	for k := range op {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(k)
		valJSON, err := json.Marshal(op[k])
		if err != nil {
			return nil, err
		}
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(valJSON)
	}
	buf.WriteByte('}')
	return []byte(buf.String()), nil
}

// AttestationEntry is a short hash followed by line ranges.
type AttestationEntry struct {
	Hash       string
	LineRanges []LineRange
}

func NewAttestationEntry(hash string, lineRanges []LineRange) AttestationEntry {
	return AttestationEntry{Hash: hash, LineRanges: lineRanges}
}

// FileAttestation is per-file attestation data.
type FileAttestation struct {
	FilePath string
	Entries  []AttestationEntry
}

func NewFileAttestation(filePath string) FileAttestation {
	return FileAttestation{FilePath: filePath}
}

func (fa *FileAttestation) AddEntry(entry AttestationEntry) {
	fa.Entries = append(fa.Entries, entry)
}

// AuthorshipLog is the complete authorship log — attestations + metadata.
type AuthorshipLog struct {
	Attestations []FileAttestation
	Metadata     AuthorshipMetadata
}

func NewAuthorshipLog() AuthorshipLog {
	return AuthorshipLog{
		Metadata: NewAuthorshipMetadata(),
	}
}

// GetOrCreateFile returns the FileAttestation for the given path, creating it if needed.
func (al *AuthorshipLog) GetOrCreateFile(filePath string) *FileAttestation {
	for i := range al.Attestations {
		if al.Attestations[i].FilePath == filePath {
			return &al.Attestations[i]
		}
	}
	al.Attestations = append(al.Attestations, NewFileAttestation(filePath))
	return &al.Attestations[len(al.Attestations)-1]
}

// SerializeToString serializes to the text+JSON authorship/3.0.0 format.
func (al *AuthorshipLog) SerializeToString() (string, error) {
	var buf strings.Builder

	for _, fa := range al.Attestations {
		filePath := fa.FilePath
		if needsQuoting(filePath) {
			filePath = fmt.Sprintf("%q", filePath)
		}
		buf.WriteString(filePath)
		buf.WriteByte('\n')

		for _, entry := range fa.Entries {
			buf.WriteString("  ")
			buf.WriteString(entry.Hash)
			buf.WriteByte(' ')
			buf.WriteString(FormatLineRanges(entry.LineRanges))
			buf.WriteByte('\n')
		}
	}

	buf.WriteString("---\n")

	jsonData, err := json.MarshalIndent(&al.Metadata, "", "  ")
	if err != nil {
		return "", err
	}
	buf.Write(jsonData)

	return buf.String(), nil
}

// DeserializeFromString parses the text+JSON authorship/3.0.0 format.
func DeserializeFromString(content string) (*AuthorshipLog, error) {
	lines := strings.Split(content, "\n")

	dividerPos := -1
	for i, line := range lines {
		if line == "---" {
			dividerPos = i
			break
		}
	}
	if dividerPos == -1 {
		return nil, aierr.NewParse("missing divider '---' in authorship log")
	}

	attestationLines := lines[:dividerPos]
	attestations, err := parseAttestationSection(attestationLines)
	if err != nil {
		return nil, err
	}

	jsonContent := strings.Join(lines[dividerPos+1:], "\n")
	var metadata AuthorshipMetadata
	if err := json.Unmarshal([]byte(jsonContent), &metadata); err != nil {
		return nil, aierr.NewJSON(err)
	}
	if metadata.Prompts == nil {
		metadata.Prompts = make(map[string]*PromptRecord)
	}

	return &AuthorshipLog{
		Attestations: attestations,
		Metadata:     metadata,
	}, nil
}

// GenerateShortHash generates a 16-character short hash from agent_id and tool.
// Deterministic: same inputs always produce the same hash.
func GenerateShortHash(agentID, tool string) string {
	combined := fmt.Sprintf("%s:%s", tool, agentID)
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash[:])[:16]
}

func parseAttestationSection(lines []string) ([]FileAttestation, error) {
	var attestations []FileAttestation
	var currentFile *FileAttestation

	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "  ") {
			entryLine := line[2:]
			spacePos := strings.Index(entryLine, " ")
			if spacePos == -1 {
				return nil, aierr.NewParse(fmt.Sprintf("invalid attestation entry format: %s", entryLine))
			}

			hash := entryLine[:spacePos]
			rangesStr := entryLine[spacePos+1:]
			lineRanges, err := ParseLineRanges(rangesStr)
			if err != nil {
				return nil, err
			}
			entry := NewAttestationEntry(hash, lineRanges)

			if currentFile == nil {
				return nil, aierr.NewParse("attestation entry found without a file path")
			}
			currentFile.AddEntry(entry)
		} else {
			if currentFile != nil && len(currentFile.Entries) > 0 {
				attestations = append(attestations, *currentFile)
			}

			filePath := line
			if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
				filePath = line[1 : len(line)-1]
			}

			fa := NewFileAttestation(filePath)
			currentFile = &fa
		}
	}

	if currentFile != nil && len(currentFile.Entries) > 0 {
		attestations = append(attestations, *currentFile)
	}

	return attestations, nil
}

func needsQuoting(path string) bool {
	return strings.ContainsAny(path, " \t\n")
}

func NewJSON(err error) *aierr.AiAttrError {
	return aierr.NewJSON(err)
}
