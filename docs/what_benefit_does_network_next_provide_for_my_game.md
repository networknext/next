<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Introduction and Overview

## The Internet doesn't care about your game

The Internet is a technological marvel, in 2023 we can watch streaming video, zoom with friends on the other side of the world and chat with friends wherever they are. Bandwidth available has never been higher, and we can download large files that just 20 years ago would have taken days. But, in the same time period, progress on quality of service (QoS) metrics such as latency, jitter and packet loss has been virtually non-existant. The Internet still provides no quality of service. Best effort delivery only.

Data from more than 25M unique players through Network Next shows that _at any time 10% of players are experiencing significantly degraded network performance_. This is true no matter where you host your game servers. In cloud, bare metal, even if that network provider claims to have the best network in the world. And each month, this bad network performance moves around, affecting more than _60% of your players at least once_.

But just how bad is it? The average packet loss around the world is 0.15%. But in many areas of the world, the average packet loss is significantly higher. Of course, averages lie and player experience is _worse_ than the average on a regular basis. Latency too is terrible. Players in Sao Paulo, Brazil experience 150ms+ additional latency as they are routed to Miami through... _New York_. Friday night, the entire Comcast backbone goes down on the East Coast, and all players on the East Coast players get an additional 100ms latency while playing until it's fixed sometime next week. A transit link is overloaded Friday night during peak play time and packet loss spikes. Players in Dubai randomly transit to game servers in Dubai, _via Frankfurt_. Korean players get higher latency than necessary when playing with players in Japan because of geopolitics, and don't get us started about players in the Middle East.

## Bad network performance reduces engagement, retention and monetization

Data from multiple games using Network Next show a consistent link between network performance and reduced engagement, retention and monetization. As latency increases, engagement, retention and monetization are reduced:

![image](https://github.com/networknext/next/assets/696656/c21bf22d-6352-4162-a085-709c4571cbe9)

Games often display a "sweet spot" or a range where latency is acceptable. For example, for this game acceleration of latency that is already below 100ms provides no benefit, however, players above 100ms should be reduced below 100ms when possible:

<img width="1112" alt="image" src="https://github.com/networknext/next/assets/696656/3c00fd6f-8825-430c-b34a-bac37c68d916">

Packet loss is different. Across all games we've worked with, the curve for packet loss is always the same. Any amount of non-trivial packet loss reduces engagement, retention and monetization:

![image](https://github.com/networknext/next/assets/696656/e224ef24-e52c-4613-82bf-576b005adaf9)

## So, what benefit does Network Next provide for my game?

Network Next platform monitors your players, and when they would otherwise experience high latency or packet loss, _we fix it_ by steering their traffic through multiple distinct routes with significantly lower latency, jitter and packet loss. Player frustration is reduced. Player engagment, retention and monetization go up.

Next: [How does it work?](how_does_it_work.md)
