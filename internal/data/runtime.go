package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Define an error that the UnmarshalJSON() method can return if we are unable to parse
// or convert the JSON string successfully.
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Declare a custom Runtime type, which has the underlying type int32
type Runtime int32

// Implement a MarshalJSON() method on the Runtime type so that it satisfies the
// json.Marshaler interface. 
// This will return a string in the format "<runtime> mins".
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the movie runtime in the required format.
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use the strconv.Quote() function on the string to wrap it in double quotes.
	// This is required so that it is valid JSON string.
	quotedJSONValue := strconv.Quote(jsonValue)

	// Convert the quoted JSON to a byte slice and return it.
	return []byte(quotedJSONValue), nil
}

// Implement a UnmarshalJSON() method on the Runtime type so that it satisfies the json.Unmarshaler interface.
// IMPORTANT: Because UnmarshalJSON() needs to modify the receiver (Runtime type), a pointer receiver is required.
// Otherwise, we'll only be modifying a copy (which is discarded when the method returns).
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect the incoming JSON value will be a string in the format "<runtime> mins".
	// First, we need to remove the surrounding double-quotes from the string.
	// If unquote fails, we return a ErrInvalidRuntimeFormat error.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number.
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Parse the string containing the number into an int32.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to Runtime type and assign it to the receiver.
	// We use the * operator to dereference the receiver (a pointer to Runtime type) and set the
	// underlying value of the pointer.
	*r = Runtime(i)
	
	return nil
}