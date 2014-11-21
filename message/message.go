package message

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Message struct {
	Type       string
	Payload    string
	Signatures []string
}

const (
	sectionBegin     = "-----BEGIN"
	sectionEndOfLine = "-----"
	sectionEnd       = "-----END"
)

func NewMessage(lines []string) *Message {
	lines = removeArmoringDelimiters(lines)

	data, err := base64.StdEncoding.DecodeString(strings.Join(lines, ""))
	if err != nil {
		fmt.Printf("err %+v\n", err)
		return nil
	}

	inflatedData, err := decompress(data)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		return nil
	}

	messageLines := strings.Split(string(inflatedData), "\n")
	message := parseMessage(messageLines)
	return &message
}

// Get rid of the header and footer. TODO I can't imagine this to be a robust
// solution but KISS.
func removeArmoringDelimiters(lines []string) []string {
	return lines[4 : len(lines)-2]
}

// Decompress zlib decompresses in and returns, if successful a byte
// array containing the decompressed output.
func decompress(in []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewBuffer(in))
	if err != nil {
		return nil, errors.New("decompression failed")
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	//log.Printf("Decompressed %d bytes from %d bytes\n", unzippedLen, len(in))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func parseMessage(lines []string) Message {
	rest, messageType, err := getType(lines)
	if err != nil {
		log.Println("err: ", err)
	}

	//FIXME Only called for the truncation of 'rest'
	rest, _, err = getHeaders(rest)
	if err != nil {
		log.Println("err: ", err)
	}

	//TODO recurse on the Payload
	rest, messagePayload, err := getPayload(rest)
	if err != nil {
		log.Println("err: ", err)
	}

	messageSignatures, err := getSignatures(rest)
	if err != nil {
		log.Println("err: ", err)
	}

	return Message{
		Type:       messageType,
		Payload:    strings.Join(messagePayload, "\n"),
		Signatures: messageSignatures,
	}
}

func getType(in []string) (rest []string, messageType string, err error) {
	rest = trimLeadingEmptyLines(in)
	if len(rest) < 2 {
		return nil, "", errors.New("Malformed document less than two lines.")
	}
	if !isCorrectBeginSectionSeperator(rest[0]) {
		return nil, "", errors.New("Header is malformed: " + rest[0])
	}

	sectionSeperator, rest := rest[0], rest[1:]
	messageType = parseSectionSeperator(sectionSeperator)
	return rest, messageType, nil
}

func getHeaders(in []string) ([]string, map[string]string, error) {
	kvPairs := make(map[string]string)
	for idx, line := range in {
		line = trim(line)
		if line == "" {
			// done
			return in[idx:], kvPairs, nil
		} else if split := strings.Split(line, ": "); len(split) == 2 {
			kvPairs[split[0]] = split[1]
		} else {
			return nil, nil, errors.New("Invalid header kv-pairs: " + line)
		}
	}

	// it should not be possible to get here if the input is well-formed
	return nil, nil, errors.New("Invalid transaction header")
}

func getPayload(in []string) ([]string, []string, error) {
	if len(in) < 2 {
		return nil, nil, errors.New("Expected at least one payload line and an end header")
	}

	var payload []string
	for idx, line := range in {
		line = trim(line)
		if isCorrectEndSectionSeperator(line) {
			// parsing is done
			return nil, payload, nil
		} else if isCorrectBeginSectionSeperator(line) {
			// there is another section, maybe a signature?
			return in[idx:], payload, nil
		} else {
			// payload
			payload = append(payload, line)
		}
	}
	// if the input is well-formed, we should never end up here
	return nil, nil, errors.New("Malformed payload")
}

func getSignatures(in []string) ([]string, error) {
	if !isCorrectBeginSectionSeperator(in[0]) {
		return []string{}, []string{}, errors.New("Expected a list of signatures")
	}

	var signatures []string
	var currentSignature string
	for _, line := range in {
		if isCorrectEndSectionSeperator(line) {
			//parsing of current signature is done
			signatures = append(signatures, currentSignature)
		} else if isCorrectBeginSectionSeperator(line) {
			//we're starting on a new signature
			currentSignature = ""
		} else if lineContainsHeader(line) {
			//NOTE: Make the header skip explicit
			continue
		} else {
			currentSignature += line
		}
	}

	//We assume there is no remainder
	return signatures, nil
}

func lineContainsHeader(line string) bool {
	return strings.Contains(line, ":")
}

func parseSectionSeperator(section string) string {
	section = strings.TrimPrefix(section, sectionBegin)
	messageType := trim(strings.TrimSuffix(section, sectionEndOfLine))
	return messageType
}

func trimLeadingEmptyLines(in []string) []string {
	for idx, line := range in {
		if line != "" {
			return in[idx:]
		}
	}
	return nil
}

func isCorrectBeginSectionSeperator(line string) bool {
	if strings.HasPrefix(line, sectionBegin) &&
		strings.HasSuffix(line, sectionEndOfLine) {
		return true
	}
	return false
}

func isCorrectEndSectionSeperator(line string) bool {
	if strings.HasPrefix(line, sectionEnd) &&
		strings.HasSuffix(line, sectionEndOfLine) {
		return true
	}
	return false
}

func trim(line string) string {
	return strings.Trim(line, " \t")
}
