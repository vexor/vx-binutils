DESTDIR?=""
SRC=$(wildcard *.go)

.PHONY: run clean install

cloud-files: $(SRC)
	gom install
	gom build

run: cloud-files
	./cloud-files

clean:
	rm -f cloud-files
	rm -rf _vendor

install: cloud-files
	mkdir -p $(DESTDIR)/usr/local/bin
	cp ./cloud-files $(DESTDIR)/usr/local/bin
