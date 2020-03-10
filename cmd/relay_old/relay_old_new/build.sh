#!/bin/bash

cd $1 && \
   premake5 gmake $2 && \
   make -j32 $3
