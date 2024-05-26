![](https://github.com/skushnerchuk/simda/actions/workflows/lint.yml/badge.svg?branch=draft)
![](https://github.com/skushnerchuk/simda/actions/workflows/test.yml/badge.svg?branch=draft)
```
  ___  ____  __  __  ____    __
 / __)(_  _)(  \/  )(  _ \  /__\   
 \__ \ _)(_  )    (  )(_) )/(__)\  
 (___/(____)(_/\/\_)(____/(__)(__)
```
### System Information Monitoring DAemon

```bash
sudo apt install -y libpcap-dev
sudo ln -f -s /usr/lib/libpcap.so /usr/lib/libpcap.so.1
```

Сборка libpcap из исходных кодов:
```bash
wget http://www.tcpdump.org/release/libpcap-1.10.4.tar.gz && \
tar xvf libpcap-1.10.4.tar.gz && \
cd libpcap-1.10.4 && \
./configure --with-pcap=linux && \
make && \
sudo make install && \
sudo mv -f ./libpcap-1.10.4/libpcap.so.1.10.4 /usr/lib/ && \
sudo ln -f -s /usr/lib/libpcap.so.1.10.4 /usr/lib/libpcap.so.1 && \
sudo ln -f -s /usr/lib/libpcap.so.1.10.4 /usr/lib/libpcap.so
```

Сборка:

```bash
go-task build
```

Запуск юнит-тестов:

```bash
go-task test
```
Запуск интеграционных тестов:

```bash
go-task integration-test
```

Демо клиента:

[demo.webm](https://github.com/skushnerchuk/simda/assets/48449889/93d1c737-5a11-4f16-932e-2db2f4b6cbbd)
