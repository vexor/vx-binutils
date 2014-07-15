DESTDIR?="$(shell pwd)/.build"

GIT_TAG     := $(shell git tag --list | head -n 1)
GIT_VERSION := $(shell echo $(GIT_TAG) | sed -e 's/v//')
ITERATION   := $(shell git rev-list $(GIT_TAG)..HEAD --count)

all:
	(cd cf && make)

clean:
	rm -rf .build
	(cd cf && make clean)

install:
	(cd cf && make DESTDIR=${DESTDIR} install)

deb:
	make install
	fpm -s dir -t deb -C $(DESTDIR) --name vx-binutils --version $(GIT_VERSION) \
		--iteration $(ITERATION) ./
