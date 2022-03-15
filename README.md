# Souv
A proof-of-concept work-in-progress web-app for surveillance camera monitoring and recording

## Getting started
### API
You will need `ffmpeg` installed locally.  
`go install && go run main.go`

### Web
`yarn && yarn dev`

### How does real time streaming from RTSP -> HLS work?
We'll get an API request that looks like something:
`POST /api/v1/stream/{camera-id}`
which will kick of a process that might be something like (first run):

- execute an ffmpeg command which takes the RTSP stream of the camera, and create a m3b8 and .ts files (ideally with -vc
  copy and -ac copy config)
- pipe these files into a well-defined location that the API can see, and make sure a GET can return both of these files
- adds the PID of the running ffmpeg process to an in-memory store with the current time, something like { camera-id:
  number, pid: number, time: time }
- return 200 if all is okay

The alternative path might be like

- check if there is an existing entry of camera-id in the in-memory store, set it to the current time.

Alongside this, there should be a process running which is frequently checking over this in-memory store and keeping it
tidy. If the time is more than 30 seconds ago, then kill the process and remove the entry from the store.

### Timing issues
It takes ffmpeg a few seconds (around 5?) before the HLS streaming really kicks off. 
The API is currently ignoring all output from the FFMPEG command, but I bet
that we should wait for it to start before we return a 200 from the initial POST
to `/api/v1/stream`.

Right now, we just wait 5 seconds before attempting to load the stream in JS