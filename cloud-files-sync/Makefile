DESTDIR?=""
SRC=main.go connection.go files.go

.PHONY: run clean install

cloud-files-sync: $(SRC)
	gom install
	gom build

run: cf
	./cf

clean:
	rm -f cloud-files-sync
	rm -rf _vendor

install: cloud-files-sync
	mkdir -p $(DESTDIR)/usr/local/bin
	cp ./cloud-files-sync $(DESTDIR)/usr/local/bin
