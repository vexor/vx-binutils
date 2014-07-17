cloud-files
===========

Синхронизация файлов с Rackspace Cloud Files

Умеет:

* загружать файлы сразу в несколько регионов
* проверяет md5 при загрузке и скачивании
* работает в 5-20 потоков, поэтому значительно быстрее [pyrax][pyrax]
* нет никаких зависимостей, один статически собранный бинарник

Не умеет:

* Работать с отличными от rackspace openstack провайдерами

Установка
=========

Сказать со страницы [releases][releases] ``deb`` или ``rpm`` пакет, зависимостей у них никаких
нет, поэтому проблем с установкой не будет

Для debian based:

    wget https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils_0.0.2-0_amd64.deb
    sudo dpkg -i vx-binutils_0.0.2-0_amd64.deb

Для rpm based:

    wget https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils-0.0.2-0.x86_64.rpm
    rpm -Uhv vx-binutils-0.0.2-0.x86_64.rpm

OSX:
    wget -O- https://github.com/vexor/vx-binutils/releases/download/v0.0.2/vx-binutils_0.0.2-2_osx_amd64.tar.gz | tar -vzxf -

Для работы потребуются 2 перменные окружения ``SDK_USERNAME`` и ``SDK_API_KEY``

    export SDK_USERNAME=<rackspace login>
    export SDK_API_KEY=<rackspace api key>

Загрузка файлов
===============

Загружать можно как весь каталог целиком так и отдельные файлы

    # синхронизирует каталог ~/packages с контейнером packages в IAD регионе,
    # эквивалента команде rsync --delete SOURCE DEST
    cloud-sync put -d -s ~/packages iad:packages

    # загрузит отдельный файл archive.tar в контейнер backup в регионе IAD
    cloud-sync put -s ~/archive.tar iad:backup

Нужный регион для загрузки указывать не обязательно, ``cloud-sync`` проверит все
регионы и найдет контейнер c указанным именем, это можно использовать для
загрузки файлов сразу в несколько регионов

    # если контейнер files есть и в IAD и в SYD регионах, то загружать будет сразу
    # в оба региона
    cloud-sync put -s ~/files files

При загрузке файлов и каталогов можно указывать prefix, который получат все загруженные файлы

    # загрузит файл backup.tar в каталог с именем 20131016/ в контейнере
    cloud-sync -s ~/backup.tar -p $(date +"%Y%m%d")/ backups


[pyrax]: https://github.com/rackspace/pyrax
[releases]: https://github.com/vexor/vx-binutils/releases
