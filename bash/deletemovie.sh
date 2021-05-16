OUTPUTMESSAGE="***** Response *****"

# Delete a movie of id 4
echo Deleting movie of id 4...
echo "$OUTPUTMESSAGE"
curl -X DELETE localhost:4000/v1/movies/4
echo

# Delete a non-existent movie
echo Deleting a non-existent movie...
echo "$OUTPUTMESSAGE"
curl -X DELETE localhost:4000/v1/movies/4
echo