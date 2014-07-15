DESTDIR?=""
SRC=main.go connection.go files.go

cf: $(SRC)
	gom install
	gom build

run: cf
	./cf

clean:
	rm -f cf
	rm -rf _vendor

install: cf
	mkdir -p $(DESTDIR)/usr/local/bin
	cp ./cf $(DESTDIR)/usr/local/bin/vx-cf
