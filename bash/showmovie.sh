OUTPUTMESSAGE="***** Response *****"

# Get an existing movie
# Remember to run createmovie.sh first
echo Get movie of id 1
echo "$OUTPUTMESSAGE"
curl -i localhost:4000/v1/movies/1
echo

# Get non-existent movie
echo Get movie of id 999999
echo "$OUTPUTMESSAGE"
curl -i localhost:4000/v1/movies/999999
echo

# Get movie of negative id
echo Get movie of id -1
echo "$OUTPUTMESSAGE"
curl -i localhost:4000/v1/movies/-1
echo

