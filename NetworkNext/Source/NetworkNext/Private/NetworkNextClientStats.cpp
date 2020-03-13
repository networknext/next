/*
    Network Next SDK 3.4.0

    Copyright © 2017 - 2020 Network Next, Inc.

	Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following
	conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions
	   and the following disclaimer in the documentation and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote
	   products derived from this software without specific prior written permission.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
	INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
	IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
	CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
	OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
	NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "NetworkNextClientStats.h"

FNetworkNextClientStats FNetworkNextClientStats::GetDisconnectedStats()
{
	FNetworkNextClientStats Stats;
	Stats.ConnectionType = ENetworkNextConnectionType::ConnectionType_Unknown;
	Stats.OnNetworkNext = false;
	Stats.DirectMinRtt = 0;
	Stats.DirectMeanRtt = 0;
	Stats.DirectMaxRtt = 0;
	Stats.DirectJitter = 0;
	Stats.DirectPacketLoss = 0;
	Stats.NetworkNextMinRtt = 0;
	Stats.NetworkNextMeanRtt = 0;
	Stats.NetworkNextMaxRtt = 0;
	Stats.NetworkNextJitter = 0;
	Stats.NetworkNextPacketLoss = 0;
	Stats.KbpsUp = 0;
	Stats.KbpsDown = 0;
	return Stats;
}