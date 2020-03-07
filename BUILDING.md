<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# How to build the Network Next SDK

## All platforms

* [Download the latest SDK source code](https://github.com/networknext/sdk/archive/master.zip)

_OR..._

* [Install git](https://git-scm.com/) _(Not required on MacOS)_

* Create an account on GitHub if you don't have one already: https://github.com

* Clone the SDK git repository:
```
git clone git@github.com:networknext/sdk.git
```

## If you are developing on Mac

* Install the XCode command line development tools if you haven't already:
```
xcode-select --install
```

* Install [brew](https://brew.sh):
```
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```

* Install dependencies via brew:
```
brew install libsodium pkg-config golang wget
```

* Install [premake5](https://premake.github.io/download.html):
```
wget https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-macosx.tar.gz && tar -zxvf premake-*.tar.gz && rm premake-*.tar.gz
```

* Generate the makefiles with premake:
```
./premake5 gmake
```

* Make everything
```
make
```

* Run the server:
```
./bin/server
```
You should see something like:
```
macbook:sdk glenn$ ./bin/server

Welcome to Network Next!

0.009219: info: server sdk version is 3.4.0
0.009238: info: server started on 127.0.0.1:32202
0.009250: info: server resolving hostname v3.networknext.com
0.110721: info: server resolved hostname to 146.148.45.204:40000
```

* Switch to another terminal and run the client:
```
./bin/client
```
You should see something like:
```
macbook:sdk glenn$ ./bin/client

Welcome to Network Next!

0.010632: info: client sdk version is 3.4.0
0.021332: info: client opened session to 127.0.0.1:32202
```

* [Go to USAGE.md for next steps](https://github.com/networknext/sdk/blob/master/USAGE.md)

## If you are developing on Windows

* [Install VS Community 2019](https://visualstudio.microsoft.com/vs/community/) if you don't already have a visual studio environment. Make sure you install or have installed the **Desktop development with C++** workload.

* Download [premake5.exe](https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-windows.zip) and put it next to `premake5.lua` in your copy of the SDK.

* Open a Command Prompt in the folder where you cloned the SDK, and run:

```
premake5 vs2019
```

* Open the generated `next.sln` solution in Visual Studio (located where you cloned the SDK). You can just double-click on `next.sln` and Visual Studio should open it.

* Under `Build` in the menu bar, click `Build Solution...` (or press `Ctrl-Shift-B`).

* In the Solution Explorer (typically on the right-hand side of the screen), right-click on `Solution 'next' (4 of 4 projects)` and click on `Set StartUp Projects...`.

* In the window that opens, select `Multiple startup projects:`.

* Next to `client`, click on (`Action`) `None` and change it to `Start`.

* Next to `server`, click on (`Action`) `None` and change it to `Start`.

* Click `OK`.

* In the toolbar click on `> Start` (the green arrow), or press `F5`.

* You should see two console windows open; one for the server, and one for the client.

* You may get a Windows Security Alert; if you do, click `Allow access` (this is the firewall access prompt since the server is starting).

* The server and client should start up, with outputs similar to the following respectively:

```
Welcome to Network Next!

0.009202: info: server sdk version is 3.4.0
0.009228: info: server datacenter is 'local'
0.009275: info: server bound to port 32202
0.009860: info: server started on 127.0.0.1:32202
0.009898: info: server resolving backend hostname 'v3.networknext.com'
0.111277: info: server resolved backend hostname to 34.69.211.185:40000

```
Welcome to Network Next!

0.027818: info: client sdk version is 3.4.0
0.028082: info: client bound to port 60994
0.054300: info: client opened session to 127.0.0.1:32202
```

* [Go to USAGE.md for next steps](https://github.com/networknext/sdk/blob/master/USAGE.md)

## If you are developing on Linux

These instructions are designed for Ubuntu 14.04-18.04.

* Install GCC:

   ```shell
   sudo apt install build-essential
   ```

* Install libsodium:

   ```shell
   sudo apt install libsodium-dev
   ```

* Install premake:

   ```shell
   curl -L -O https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-src.zip && \
   unzip premake-5.0.0-alpha14-src.zip && \
   cd premake-5.0.0-alpha14 && \
   cd build/gmake.unix && \
   make && \
   sudo mv ../../bin/release/premake5 /usr/local/bin && \
   cd ../../../ && \
   rm -rf premake-*
   ```

* If you want to build 32-bit:
   ```shell
   sudo dpkg --add-architecture i386
   sudo apt update
   sudo apt install gcc-multilib g++-multilib
   ```

* Choose a build configuration:
   - `debug_static32`
   - `debug_shared32`
   - `debug_static64`
   - `debug_shared64`
   - `release_static32`
   - `release_shared32`
   - `release_static64`
   - `release_shared64`

   Then:

    ```shell
    premake5 gmake
    make config=<config>
    ./bin/server &
    ./bin/client
    ```

* [Go to USAGE.md for next steps](https://github.com/networknext/sdk/blob/master/USAGE.md)

## If you are developing on Docker

 ```shell
 docker build -t networknext .
 docker run -ti --rm networknext bin/server
 docker run -ti --rm networknext bin/client
 ```

* [Go to USAGE.md for next steps](https://github.com/networknext/sdk/blob/master/USAGE.md)
