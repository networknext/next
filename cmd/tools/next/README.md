<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

<h1>Network Next Operator Tool</h1>

Examples:

Select localhost environment (default):

```
next select localhost
```

Select development environment:

```
next select dev.networknext.com
```

Select production environment:

```
next select prod.networknext.com
```

Print current environment (stored in env.txt):

```
next env
```

List all relays names:

```
next relays
```

List all relay names matching 'i3d' substring:

```
next relays i3d
```

List all datacenters:

```
next datacenters
```

List all datacenters matching 'i3d' substring:

```
next datacenters i3d
```

SSH into a relay by name:

```
next ssh i3d.rotterdam.a
```

Force sync all environment data to local cache (env.json):

```
next sync
```

Clean all cached data:

```
next clean
```
