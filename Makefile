twitter-follower-rank: build gen-docs

build:
	go build -o .

gen-docs:
	./gen-docs.sh
