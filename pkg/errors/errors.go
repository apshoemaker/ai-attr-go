package errors

import "fmt"

type ErrorKind int

const (
	ErrIO ErrorKind = iota
	ErrJSON
	ErrGit
	ErrNotAGitRepo
	ErrParse
	ErrOther
)

type AiAttrError struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *AiAttrError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AiAttrError) Unwrap() error {
	return e.Err
}

func NewIO(err error) *AiAttrError {
	return &AiAttrError{Kind: ErrIO, Message: "IO error", Err: err}
}

func NewJSON(err error) *AiAttrError {
	return &AiAttrError{Kind: ErrJSON, Message: "JSON error", Err: err}
}

func NewGit(msg string) *AiAttrError {
	return &AiAttrError{Kind: ErrGit, Message: msg}
}

func NewNotAGitRepo() *AiAttrError {
	return &AiAttrError{Kind: ErrNotAGitRepo, Message: "not a git repository"}
}

func NewParse(msg string) *AiAttrError {
	return &AiAttrError{Kind: ErrParse, Message: msg}
}

func NewOther(msg string) *AiAttrError {
	return &AiAttrError{Kind: ErrOther, Message: msg}
}
