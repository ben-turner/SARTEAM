package radiotracker

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type MessageType uint
type MessageFormat string
type parseConfig struct {
	format      MessageFormat
	messageType int
	lat         int
	lon         int
	radioID     int
}

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeActiveGPS
	MessageTypeStaleGPS

	MessageFormatPKLDS MessageFormat = "PKLDS"
	MessageFormatGPRMC MessageFormat = "GPRMC"
)

var (
	parseConfigPKLDS = parseConfig{
		format:      MessageFormatPKLDS,
		messageType: 2,
		lat:         3,
		lon:         5,
		radioID:     13,
	}

	parseConfigGPRMC = parseConfig{
		format:      MessageFormatGPRMC,
		messageType: 2,
		lat:         3,
		lon:         5,
		radioID:     -4, // Each line is actually multiple messages and we can parse the radio ID from the previous one.
	}
)

var (
	ErrUnknownMessageFormat = errors.New("unknown message format")
	ErrInvalidCoordinate    = errors.New("invalid coordinate")
)

type Message struct {
	Format MessageFormat
	Type   MessageType

	Latitude  float64
	Longitude float64

	Timestamp time.Time

	RadioID string
}

func (r *Message) GetBytes() []byte {
	return []byte{}
}

// parseCoordPart parses a coordinate part (either lat or lon) into a float64.
// Coordinate parts are assumed to be the concatenation of the degrees and
// decimal minutes, without a separator, but having a decimal point in the
// minutes part. This is perhaps more clearly explained by the following regex:
//
// `(?P<degrees>[0-9]+)(?P<minutes>[0-9]{2}\.[0-9]+)`
func parseCoordPart(part string) (float64, error) {
	degreeDigits := strings.Index(part, ".") - 2
	if degreeDigits < 0 {
		return 0, ErrInvalidCoordinate
	}

	degreesStr := part[:degreeDigits]
	minutesStr := part[degreeDigits:]

	degrees, err := strconv.ParseFloat(degreesStr, 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(minutesStr, 64)
	if err != nil {
		return 0, err
	}

	return degrees + minutes/60, nil
}

func parseCoordinates(lat, lon string) (float64, float64, error) {
	outLat, err := parseCoordPart(lat)
	if err != nil {
		return 0, 0, err
	}

	outLon, err := parseCoordPart(lon)
	if err != nil {
		return 0, 0, err
	}

	return outLat, outLon, nil
}

func parseMessage(words []string, offset int, config parseConfig, ts time.Time) (*Message, error) {
	msg := &Message{
		Format:    config.format,
		RadioID:   words[offset+config.radioID],
		Timestamp: ts,
	}

	switch words[offset+config.messageType] {
	case "A":
		msg.Type = MessageTypeActiveGPS
	case "V":
		msg.Type = MessageTypeStaleGPS
	default:
		msg.Type = MessageTypeUnknown
		return msg, nil
	}

	lat, lon, err := parseCoordinates(words[offset+config.lat], words[offset+config.lon])
	if err != nil {
		return nil, err
	}

	msg.Latitude = lat
	msg.Longitude = lon
	return msg, nil
}

func ParseMessage(msg *rawMessage) (*Message, error) {
	words := strings.FieldsFunc(msg.raw, func(r rune) bool { return r == ',' || r == ' ' || r == '\u0002' })
	for i, word := range words {
		switch word {
		case "$PKLDS":
			return parseMessage(words, i, parseConfigPKLDS, msg.ts)
		case "$GPRMC":
			return parseMessage(words, i, parseConfigGPRMC, msg.ts)
		}
	}

	return nil, ErrUnknownMessageFormat
}
