FROM golang:1.19
WORKDIR /
EXPOSE 8080
RUN git clone https://github.com/nitesh-repos/AccuKnox.git

# RUN cd /AccuKnox/app
# RUN go mod init accu_knox
# RUN go mod tidy
# RUN go build -o main main.go
# CMD main
CMD while true; do sleep 1; done


# docker build -t notes_api_img .
# docker run -itd --name notes_api -p8080:80 notes_api_img
# docker exec -it notes_api sh

