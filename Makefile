.PHONY: build clean daemon

daemon:
	docker desktop start

build:
	docker build -t eeestrelok/gns3-golang:latest . && \
	docker push eeestrelok/gns3-golang:latest

clean:
	docker rmi $(FULL_IMAGE)