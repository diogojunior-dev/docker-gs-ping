# A simple microsercive in Go packaged into a container image.

FROM golang:1.20

# Create a work directory, form now on I can use relative paths
WORKDIR /app

# Copy the module files before downloading the dependencies
COPY go.* ./

# Download de dependencies 
RUN go mod download

# Copy the souce code into the umage
COPY *.go ./

# Run the command to compile the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

#Documente what port the application is going to listen on by default
EXPOSE 8080

# Run 
CMD ["/docker-gs-ping"]