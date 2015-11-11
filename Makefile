# Thanks for inspiration
# - https://gist.github.com/yanatan16/2951128
# - https://github.com/mackerelio/mackerel-agent/blob/master/Makefile

# Followings are ommited from 'deps' because of need of sudo 
# sudo gem install fpm

FPM = fpm
REPO="github.com/shiguredo/fuji/cmd/fuji"
TEST_LIST := tests

TAG=0.3.0
REV=`git rev-parse HEAD | cut -c1-7`
ARTIFACTS=downloads
BUILDDIR=build
LDFLAGS=-ldflags "-X main.version=${TAG}-${REV}"

ALL_LIST = $(TEST_LIST)

SUDOPATH=${PATH}

all: build test raspi raspi2 edison armadillo

fmt:
	go fmt ./...

doc: fmt
#	golint ./...
	go build github.com/shiguredo/fuji/cmd/fuji
	godoc github.com/shiguredo/fuji

build: fmt deps
#	golint ./...
	go build
	go build $(LDFLAGS) github.com/shiguredo/fuji/cmd/fuji

###### target by architecture

arm5: deps
	GOARM=5 gox $(LDFLAGS) -os="linux" -arch="arm" -output=$(BUILDDIR)/arm5/fuji/fuji-gw $(REPO)
	cp -p packages/config.simple.toml.example $(BUILDDIR)/arm5/fuji/config.toml.example
	cd $(BUILDDIR)/arm5/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_arm5.tar.gz fuji
	echo 'linux arm5 build completed'

arm6: deps
	GOARM=6 gox $(LDFLAGS) -os="linux" -arch="arm" -output=$(BUILDDIR)/arm6/fuji/fuji-gw $(REPO)
	cp -p packages/config.simple.toml.example $(BUILDDIR)/arm6/fuji/config.toml.example
	cd $(BUILDDIR)/arm6/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_arm6.tar.gz fuji
	echo 'linux arm6 build completed'

arm7: deps
	GOARM=7 gox $(LDFLAGS) -os="linux" -arch="arm" -output=$(BUILDDIR)/arm7/fuji/fuji-gw $(REPO)
	cp -p packages/config.simple.toml.example $(BUILDDIR)/arm7/fuji/config.toml.example
	cd $(BUILDDIR)/arm7/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_arm7.tar.gz fuji
	echo 'linux arm7 build completed'

linux_386: deps
	gox $(LDFLAGS) -os="linux" -arch="386" -output=$(BUILDDIR)/linux_386/fuji/fuji-gw $(REPO)
	cp -p packages/config.simple.toml.example $(BUILDDIR)/linux_386/fuji/config.toml.example
	cd $(BUILDDIR)/linux_386/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_linux_386.tar.gz fuji
	echo 'linux 386 build completed'

linux_amd64: deps
	gox $(LDFLAGS) -os="linux" -arch="amd64" -output=$(BUILDDIR)/linux_amd64/fuji/fuji-gw $(REPO)
	cp -p packages/config.simple.toml.example $(BUILDDIR)/linux_amd64/fuji/config.toml.example
	cd $(BUILDDIR)/linux_amd64/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_linux_amd64.tar.gz fuji
	echo 'linux amd64 build completed'

# freebsd_amd64: deps
# 	gox $(LDFLAGS) -os="freebsd" -arch="amd64" -output=$(BUILDDIR)/freebsd_amd64/fuji/fuji-gw $(REPO)
# 	cp -p packages/config.simple.toml.example $(BUILDDIR)/freebsd_amd64/fuji/config.toml.example
# 	cd $(BUILDDIR)/freebsd_amd64/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_freebsd_amd64.tar.gz fuji
# 	echo 'freebsd amd64 build completed'

# darwin_amd64: deps
# 	gox $(LDFLAGS) -os="darwin" -arch="amd64" -output=$(BUILDDIR)/darwin/fuji/fuji-gw $(REPO)
# 	cp -p packages/config.simple.toml.example $(BUILDDIR)/darwin/fuji/config.toml.example
# 	cd $(BUILDDIR)/darwin/ && tar zcvf ../../$(ARTIFACTS)/fuji-gw_$(TAG)_darwin_amd64.tar.gz fuji
# 	echo 'darwin amd64 build completed'

###### packaging

raspi: deps arm6
	gem install fpm
	if [ -d $(BUILDDIR)/packages_raspi ] ; then rm -rf $(BUILDDIR)/packages_raspi ; fi
	mkdir -p $(BUILDDIR)/packages_raspi/usr/local/bin
	mkdir -p $(BUILDDIR)/packages_raspi/etc/fuji-gw
	mkdir -p $(BUILDDIR)/packages_raspi/etc/init.d
	cp -p $(BUILDDIR)/arm6/fuji/fuji-gw $(BUILDDIR)/packages_raspi/usr/local/bin/fuji-gw
	cp -p packages/config.simple.toml.example $(BUILDDIR)/packages_raspi/etc/fuji-gw/config.toml
	cp -p packages/fuji-gw.init $(BUILDDIR)/packages_raspi/etc/init.d/
	cd $(BUILDDIR)/packages_raspi; $(FPM) -s dir -t deb -a armhf -n fuji-gw -v $(TAG) -p ../../$(ARTIFACTS) --deb-init ./etc/init.d/fuji-gw.init .
	mv $(ARTIFACTS)/fuji-gw_$(TAG)_armhf.deb $(ARTIFACTS)/fuji-gw_$(TAG)_raspi_arm6.deb

raspi2: deps arm7
	gem install fpm
	if [ -d $(BUILDDIR)/packages_raspi2 ] ; then rm -rf $(BUILDDIR)/packages_raspi2 ; fi
	mkdir -p $(BUILDDIR)/packages_raspi2/usr/local/bin
	mkdir -p $(BUILDDIR)/packages_raspi2/etc/fuji-gw
	cp -p $(BUILDDIR)/arm7/fuji/fuji-gw $(BUILDDIR)/packages_raspi2/usr/local/bin/fuji-gw
	cp -p packages/config.simple.toml.example $(BUILDDIR)/packages_raspi2/etc/fuji-gw/config.toml
	cd $(BUILDDIR)/packages_raspi2; $(FPM) -s dir -t deb -a armhf -n fuji-gw -v $(TAG) -p ../../$(ARTIFACTS) .
	mv $(ARTIFACTS)/fuji-gw_$(TAG)_armhf.deb $(ARTIFACTS)/fuji-gw_$(TAG)_raspi2_arm7.deb

armadillo: deps arm5
	gem install fpm
	if [ -d $(BUILDDIR)/packages_armadillo ] ; then rm -rf $(BUILDDIR)/packages_armadillo ; fi
	mkdir -p $(BUILDDIR)/packages_armadillo/usr/local/bin
	mkdir -p $(BUILDDIR)/packages_armadillo/etc/fuji-gw
	cp -p $(BUILDDIR)/arm5/fuji/fuji-gw $(BUILDDIR)/packages_armadillo/usr/local/bin/fuji-gw
	cp -p packages/config.simple.toml.example $(BUILDDIR)/packages_armadillo/etc/fuji-gw/config.toml
	cd $(BUILDDIR)/packages_armadillo; $(FPM) -s dir -t deb -a armle -n fuji-gw -v $(TAG) -p .. .

edison: linux_386
	if [ ! -d opkg-utils ] ; then git clone http://git.yoctoproject.org/git/opkg-utils ; fi
	if [ -d $(BUILDDIR)/packages_edison ] ; then rm -rf $(BUILDDIR)/packages_edison ; fi
	mkdir -p $(BUILDDIR)/packages_edison/usr/local/bin
	mkdir -p $(BUILDDIR)/packages_edison/etc/fuji-gw

	cp -p $(BUILDDIR)/linux_386/fuji/fuji-gw $(BUILDDIR)/packages_edison/usr/local/bin/fuji-gw
	cp -p packages/config.simple.toml.example $(BUILDDIR)/packages_edison/etc/fuji-gw/config.toml

	mkdir -p $(BUILDDIR)/packages_edison/CONTROL
	sed -e 's/FUJI_GIT_TAG/$(TAG)/' packages/opkg_files/control.in > packages/opkg_files/control
	cp packages/opkg_files/control $(BUILDDIR)/packages_edison/CONTROL
	cd packages/opkg_files && tar czf control.tar.gz control
	cd $(BUILDDIR)/packages_edison/ && sudo ../../opkg-utils/opkg-build -o root -g root . /tmp
	mv /tmp/fuji-gw_$(TAG)_edison.ipk $(ARTIFACTS)/fuji-gw_$(TAG)_edison_386.ipk

linux_amd64_deb: linux_amd64
	gem install fpm
	if [ -d $(BUILDDIR)/packages_linux_amd64 ] ; then rm -rf $(BUILDDIR)/packages_linux_amd64 ; fi
	mkdir -p $(BUILDDIR)/packages_linux_amd64/usr/local/bin
	mkdir -p $(BUILDDIR)/packages_linux_amd64/etc/fuji-gw
	cp -p $(BUILDDIR)/linux_amd64/fuji/fuji-gw $(BUILDDIR)/packages_linux_amd64/usr/local/bin/fuji-gw
	cp -p packages/config.simple.toml.example $(BUILDDIR)/packages_linux_amd64/etc/fuji-gw/config.toml
	cd $(BUILDDIR)/packages_linux_amd64; $(FPM) -s dir -t deb -a x86_64 -n fuji-gw -v $(TAG) -p ../../$(ARTIFACTS) .


test: $(ALL_LIST)
	go get golang.org/x/tools/cmd/cover
	go test -coverpkg github.com/shiguredo/fuji ./...

deps:
	if [ -d $(ARTIFACTS) ] ; then rm -rf $(ARTIFACTS) ; fi
	if [ -d $(BUILDDIR) ] ; then rm -rf $(BUILDDIR) ; fi
	mkdir -p $(ARTIFACTS)
	mkdir -p $(BUILDDIR)

	go get -d -v -t ./...
	go get -u github.com/mitchellh/gox
	go get github.com/kr/pty
#	go get -u github.com/golang/lint/golint
	godep restore

clean:
	rm -f fuji
	rm -rf build downloads opkg-utils
	rm -rf packages/opkg_files/control packages/opkg_files/control.tar.gz
