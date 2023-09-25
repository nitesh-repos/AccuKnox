FROM golang:1.19
WORKDIR /app
EXPOSE 8081
RUN git clone https://github.com/nitesh-repos/AccuKnox.git

RUN mv AccuKnox/rest_api rest_api
RUN rm -rf AccuKnox
RUN cd rest_api
RUN go mod init rest_api
RUN go mod tidy
RUN go build -o main main.go
CMD ./main

# CMD while true; do sleep 1; done
# docker build -t notes_api_img .; docker run -itd --name notes_api -p8081:80 notes_api_img; docker exec -it notes_api sh
# docker commit notes_api_cont
