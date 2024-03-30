<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Portal user guide

The portal is where you can see what's going on with your players in real-time.

It lives at:

* https://portal-dev.[yourdomain.com] for your development environment
* https://portal-staging.[yourdomain.com] for your staging environment (load tests)
* https://portal.[yourdomain.com] for your production environment

## 1. Session Counts

At the top of each page, you can always see the current session counts for your system:

<img width="1468" alt="image" src="https://github.com/networknext/next/assets/696656/e37e9740-d810-416a-b6ec-3d9b796eb442">

These counts are updated once per-minute. They are equivalent to the number of unique session ids seen across one minute, so if you have an instantaneous player count calculated somewhere else, you will notice that these numbers are a bit higher than that.

## 2. Sessions Page

The default page is the sessions page. Here we see the top player sessions, in order from largest acceleration improvement to least:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/90b98369-7d2b-4473-9b17-4301a7a8501e">

If acceleration is disabled, then you will see player sessions sorted in order of highest latency to lowest latency.

One hundred sessions are shown on each page. To navigate left and right across all sessions, use the 1 and 2 keys to navigate left and right. On mobile layouts, arrows icons are shown in the navbar to let you navigate without a keyboard.

## 3. Session detail page

If you click on the session id for a session:

![image](https://github.com/networknext/next/assets/696656/0d3dc656-0eba-4ca9-ab66-e731b0cc6483)

You go to the session detail page for that session:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/841b5bd0-21ad-4915-a050-7c01848ee9c4">

You can also get to here by crafting a URL with the session id in hex:

`https://portal-dev.virtualgo.net/session/53144bb8430e3ac2`

Under the session detail page you can immediately see a graph of latency over time for that session:

<img width="1002" alt="image" src="https://github.com/networknext/next/assets/696656/7f99ffc5-cfd1-40b9-a882-28baf59a8b01">

The blue line is the non-accelerated direct route round trip time (RTT) from the client to the server and back in milliseconds. The green line is the accelerated round trip time in milliseconds. The orange line is the conservative predicted accelerated round trip time calculated by the route optimization system.

Below we can see jitter, also measured in milliseconds. Jitter measures time variance of packet delivery in milliseconds, in other words, packets that arrive later than they should:

<img width="1002" alt="image" src="https://github.com/networknext/next/assets/696656/2d28e15a-dfa6-46a4-b2c1-6de887231424">

The purple real jitter value is the most important to consider. It is the average difference in packet delivery time measured for a full round trip from the client to the server and back. It is affected by your server tick rate, so if you send packets from the client to server at some high rate like 60 packets per-second, but the server replies at only 10HZ tick intervals, you will see that reflected in the real jitter value.

The blue and green jitter values represent estimates of jitter, excluding server tick rate effects with a more conservative metric. These values will be zero on perfect connections, and usually are less than 5-10 milliseconds on good wi-fi connections. If the player has a poor wi-fi connection you will often see these values spike to 50-100ms. These players should be advised to play over a wired connection for the best experience.

Packet loss is the percentage of your game packets that are lost in the last 10 seconds. It is common for players with excellent connections to have very little, or no packet loss. Some packet loss tends to happen more frequently on bad wi-fi, and of course players in less populated areas tend to have higher packet loss. The average packet loss overall around the world is 0.15%.

<img width="1026" alt="image" src="https://github.com/networknext/next/assets/696656/ebdf5c09-638d-4bc5-a0f1-671a2230d7be">

Out of order packets is the percentage of packets that are received out of sequence order in the last 10 seconds. On most connections this will be 0%, but on poor wi-fi connections or bad internet connections out of order can be non-zero:

<img width="1026" alt="image" src="https://github.com/networknext/next/assets/696656/caeedc68-5209-435f-9044-e8553231f7ed">

Bandwidth is shown over time in the client to server (up), and server to client (down) directions separately - because most games tend to have asymmetric bandwidth usage:

<img width="1026" alt="image" src="https://github.com/networknext/next/assets/696656/42d6fe6b-5617-464f-bc4b-70b419435681">

The blue line is the bandwidth sent along the unaccelerated codepath, and the green is the bandwidth sent along the accelerated codepath in kilobits per-second. Kilo-bits. Not bytes. Note that Network Next by default sends packets across both the unaccelerated and accelerated codepaths at the same time. This is called "multipath" and it helps to reduce packet loss significantly.

On the right side of the session detail page you will see some useful summary information about the session:

<img width="1468" alt="image" src="https://github.com/networknext/next/assets/696656/b463d1b3-918a-4018-973f-3f511289026b">

Many of these values are clickable. For example, clicking on the user hash goes to a list of sessions this user has played recently, so you can see this user's history of network performance. Clicking on the server will take you to the list of sessions connected to that server currently, so you can look at other players on the server.

Below the summary data, you will see the current route from the client to the server. If the session is accelerated, the relays that traffic is being sent through will be shown here, so you know what route your packets are taking:

<img width="1468" alt="image" src="https://github.com/networknext/next/assets/696656/0e656ba5-95c0-42b9-b1fd-6ada402dc646">

You can even click on each relay in the route, and you will be taken to a detail page for that relay.

## 4. Servers page

Servers are your game servers that are currently running with the Network Next SDK.

You can view the list of servers by clicking on "Servers" in the top menu. If you need to go left/right in the list, you use the keys 1/2 to scroll.

![image](https://github.com/networknext/next/assets/696656/6fc898b9-c611-4e27-b2be-85eeca7f375f)

You can click on the server IP address and port to view details for that server:

<img width="1467" alt="image" src="https://github.com/networknext/next/assets/696656/545be585-9010-47bb-afe7-0bc3e329ed6e">

You can also get server details directly by crafting a URL with the server IP address and port, like this:

`https://portal-dev.virtualgo.net/server/35.239.236.74:32834`

On the server page you can see the set of sessions connected to the server at the current time. You can click through on any of these sessions to drill into each session in detail, or click on the user hash to see what sessions.

## 5. Relays page

Relays are software routers that traffic is passed through when Network Next performs acceleration. The system is only as good as your relay fleet. A typical production relay fleet will have at least 10 relays in each major city from different providers, to get a variety networks and ISP connections.

The list of relays is shown under the "Relays" item in the top menu:

![image](https://github.com/networknext/next/assets/696656/3602af83-6e5f-48ff-a4ed-d3540ac39980)

You can click on the relay name to drill in and see graphs over time for this relay:

<img width="1468" alt="image" src="https://github.com/networknext/next/assets/696656/67db43f5-28ae-4d96-89e6-72e85398ce4c">

## 6. Datacenters page

Datacenters are named physical locations that relays are attached to. Every relay has just one datacenter that it is in, but datacenters can have multiple relays in them. Datacenters have a seller, eg. "Google", and a physical location: latitude and longitude.

You can see the list of all datacenters currently defined in your system by going to the datacenters page:

<img width="1467" alt="image" src="https://github.com/networknext/next/assets/696656/923d4db3-b204-48b0-82ff-37f9742c1cf4">

## 7. Buyers page

Buyers correspond to games using your Network Next system. Games "buy" network transit from sellers (Google, AWS, Linode etc.)

<img width="1469" alt="image" src="https://github.com/networknext/next/assets/696656/40985282-44e2-4b09-ad5a-9628150b3378">

You can click on any buyer and see some graphs such as the buyer's total sessions, accelerated sessions, accelerated % and server count over the last hour.

## 8. Sellers page

Sellers are companies that supply relays. Each relay is in a datacenter, and each datacenter belongs to a seller.

The sellers page shows you the set of sellers defined in your system:

<img width="1467" alt="image" src="https://github.com/networknext/next/assets/696656/5177a491-75b1-4de4-9814-d2456e8f04d5">

If you click on a seller, it will give you a list of relays belonging to that seller.

## 9. Admin page

The admin page shows important graphs to check the health of the system as a whole:

<img width="1470" alt="image" src="https://github.com/networknext/next/assets/696656/177cfd36-abe2-442a-9e1d-084d472e7ff1">

1. Total sessions (total sessions across all buyers)
2. Accelerated sessions (accelerated sessions across all buyers)
3. Accelerated % (percentage of acceleration over time across all buyers. naturally fluctuates somewhat)
4. Server count (tracks the number of servers across all buyers)
5. Active relays (the number of relays in your relay fleet)
6. Retries (number of server to backend requests from the SDK that needed to be retried. early indicator of server backend overload)
7. Fallback to direct (fallback to direct is generally bad. sessions lose acceleration and won't come back to talk to the server backend again until that player starts the next play session. usually a consequence of server backend overload).
8. Total routes (the total number of routes in the route matrix used for route planning. fluctuates)
9. Optimize time in seconds (if this frequently goes over 1 second, then you need to increase the number of cores on the relay backend!)

That's it! I hope you get good visibility on your player's network experience with the portal.

[Go back to main documentation](../README.md)
