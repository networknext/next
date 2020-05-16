
Building the SDK
================

Official binaries for Windows and consoles have already been built for you in the "lib" directory. 

To build the SDK yourself, and the example programs, see the instructions below:

Windows
-------

Install *premake5* from https://premake.github.io/download.html

Generate a visual studio solution via premake:

.. code-block::

    $ premake5 vs2017

Open the generated solution file and build all.

Native Visual Studio solutions are also available in the "build" directory.

Mac
---

Make sure the XCode command line tools are installed:

.. code-block::

	$ xcode-select --install

Install *premake5* from https://premake.github.io/download.html

Generate makefiles by running premake at the root directory of the SDK:

.. code-block::

    $ premake5 gmake

Build the SDK:

.. code-block::

	$ make

Run the unit tests:

.. code-block::

	$ ./bin/test

Linux
-----

Make sure the build essential package is installed:

.. code-block::

	$ sudo apt install build-essential

Install *premake5* from https://premake.github.io/download.html

Generate makefiles by running premake at the root directory of the SDK:

.. code-block::

    $ premake5 gmake

Build the SDK:

.. code-block::

	$ make

Run the unit tests:

.. code-block::

	$ ./bin/test
