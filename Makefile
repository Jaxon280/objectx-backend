SRCS = *.go
PROGRAM = objectx

build: $(SRCS)
	GOOS=darwin go build -o $(PROGRAM) $(SRCS)
	mv $(PROGRAM) /usr/local/bin/

clean:
	rm -f *.db

.PHONY: build clean
