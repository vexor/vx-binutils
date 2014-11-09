cloud-files
===========

Synchronising files with Rackspage Cloud Files

Can do:

* upload files simultaneously in separate regions
* verify MD5 when uploading and downloading
* utilize 5-20 threads, so faster than [pyrax][pyrax]
* work without dependencies, using a single statically built binary


Cannot do:

* Work with openstack providers, other than rackspace

Setting up
=========

Download ``deb`` or ``rpm`` package from the [releases][releases] page. There are no dependencies
so you shouldn't have any issues with installation.


For debian based:

    wget https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils_0.0.2-0_amd64.deb
    sudo dpkg -i vx-binutils_0.0.2-0_amd64.deb

For rpm based:

    wget https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils-0.0.2-0.x86_64.rpm
    rpm -Uhv vx-binutils-0.0.2-0.x86_64.rpm

OSX:
    wget -O- https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils_0.0.2-2_osx_amd64.tar.gz | tar -vzxf -

Two environmental variables are required:  ``SDK_USERNAME`` and ``SDK_API_KEY``

    export SDK_USERNAME=<rackspace login>
    export SDK_API_KEY=<rackspace api key>

Uploading files
===============

It is possible to upload either specific files or entire directories

    # synchronizes the ~/packages directory with packages container in IAD region,
    # which is equivalent to running rsync --delete SOURCE DEST
    cloud-sync put -d -s ~/packages iad:packages

    # uploads archive.tar file into backup container in IAD region
    cloud-sync put -s ~/archive.tar iad:backup

You don't have to specify the required region when uploading, ``cloud-sync`` will check all regions
and will find the container with given name, so it is possible to upload multiple files into several regions
at the same time.


  
    # if the files container exists in both IAD and SYD regions, it will start uploading to both of them
    cloud-sync put -s ~/files files

You can add prefix when uploading, and this prefix will be used for all the uploaded file names.

    
    # uploads backup.tar file into the 20131016/ directory in the container
    cloud-sync -s ~/backup.tar -p $(date +"%Y%m%d")/ backups


[pyrax]: https://github.com/rackspace/pyrax
[releases]: https://github.com/vexor/vx-binutils/releases
