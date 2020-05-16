
Building the SDK
================

Prebuilt binaries for Windows and console have already been built for you in the "lib" directory.

To build directly from source, see the instructions below:

MacOS
-----

Make sure the XCode command line tools are installed:

.. code-block::

	$ xcode-select --install

Install *premake5* from https://premake.github.io/download.html

Generate makefiles via premake at the root at the SDK directory:

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

Make sure the C++ build essentials package is installed:

.. code-block::

	$ sudo apt install build-essential

Install *premake5* from https://premake.github.io/download.html

Generate makefiles via premake at the root of the SDK directory:

.. code-block::

    $ premake5 gmake

Build the SDK:

.. code-block::

	$ make

Run the unit tests:

.. code-block::

	$ ./bin/test

Windows
-------

Install *premake5* from https://premake.github.io/download.html

Generate the visual studio solution via premake:

.. code-block::

    $ premake5 vs2017

Open the generated solution file and build all.
