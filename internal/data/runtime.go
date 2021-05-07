package data

import (
	"fmt"
	"strconv"
)

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