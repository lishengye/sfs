all: sfs sfc

sfs:
	go build -o ./output/sfs  ./cmd/server
	cp config/sfs.json  output/

sfc:
	go build -o ./output/sfc  ./cmd/client

clean:
	rm -f output/*