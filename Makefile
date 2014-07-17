DESTDIR?="$(shell pwd)/.build"

GIT_TAG     := $(shell git tag --list | sort -r | head -n 1)
GIT_VERSION := $(shell echo $(GIT_TAG) | sed -e 's/v//')
ITERATION   := $(shell git rev-list $(GIT_TAG)..HEAD --count)

.PHONY: build clean install deb

build:
	(cd cloud-files && make)

clean:
	rm -rf .build
	rm -rf *.deb
	(cd cloud-files && make clean)

install: build
	(cd cloud-files && make DESTDIR=${DESTDIR} install)

deb:
	make install
	fpm -s dir -t deb -C $(DESTDIR) --name vx-binutils --version $(GIT_VERSION) \
		--iteration $(ITERATION) ./

rpm:
	make install
	fpm -s dir -t rpm -C $(DESTDIR) --name vx-binutils --version $(GIT_VERSION) \
		--iteration $(ITERATION) ./
