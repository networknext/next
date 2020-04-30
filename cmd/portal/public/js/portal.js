/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg';

var userInfo = null;

var accountsTable = null;
var pubKeyInput = null;
var relaysTable = null;
var sessionsTable = null;

function startApp() {
	Promise.all([
		/**
		 * QUESTION: Instead of grabbing the user here can we use the token to then go off and get everything from the backend?
		 * TODO:	 There are 3 different promises going off to get user details. There should be a better way to do this
		 */
		loginClient.getUser(),
		loginClient.getTokenSilently()
	]).then((response) => {
		userInfo = {
			email: response[0].email,
			name: response[0].name,
			nickname: response[0].nickname,
			userId: response[0].sub,
			token: response[1]
		};
		JSONRPCClient
			.call('AuthService.UserRoles', {user_id: userInfo.userId})
			.then((response) => {
				userInfo.roles = response.roles;

				document.getElementById("app").style.display = 'block';
				MapHandler
					.initMap()
					.then((response) => {
						console.log("Map init successful");
					})
					.catch((error) => {
						console.log("Map init unsuccessful: " + error);
					});
			})
			.catch((e) => {
				console.log("Something went wrong with getting the user roles");
				console.log(e);
			});
		})
		.catch((e) => {
			console.log("Something went wrong fetching user details or user token");
			console.log(e);
		});
}

function changePage(page) {
	let account = document.getElementById("account-workspace");
	let map = document.getElementById("map-workspace");
	let relay = document.getElementById("relay-workspace");
	let session = document.getElementById("session-workspace");
	let title = document.getElementById("workspace-title");
	let users = document.getElementById("users-workspace");

	let mapLink = document.getElementById("home-link");
	let relayLink = document.getElementById("relay-link");
	let sessionLink = document.getElementById("session-link");
	let usersLink = document.getElementById("users-link");

	account.style.display = 'none';
	map.style.display = 'none';
	relay.style.display = 'none';
	session.style.display = 'none';
	users.style.display = 'none';

	mapLink.classList.remove("active");
	relayLink.classList.remove("active");
	sessionLink.classList.remove("active");
	usersLink.classList.remove("active");

	switch (page) {
		case 'account':
			loadAccountPage();
			account.style.display = 'block';
			title.textContent = 'Account Details';
			break;
		case 'relay':
			loadRelayPage();
			relay.style.display = 'block';
			relayLink.classList.add("active");
			title.textContent = 'Relays Table';
			break;
		case 'sessions':
			loadSessionPage();
			session.style.display = 'block';
			sessionLink.classList.add("active");
			title.textContent = 'Session Table';
			break;
		case 'users':
			loadUsersPage();
			users.style.display = 'block';
			usersLink.classList.add("active");
			title.textContent = 'User Table';
			break;
		default:
			map.style.display = 'block';
			mapLink.classList.add("active");
			title.textContent = 'Session Map';
	}
}

function changeAccountPage(page) {
	let config = document.getElementById("config");
	let accounts = document.getElementById("accounts");
	let newUser = document.getElementById("new-user");
	let newUserButton = document.getElementById("new-user-button");

	let accountsLink = document.getElementById("accounts-link");
	let configLink = document.getElementById("config-link");

	accounts.style.display = 'none';
	config.style.display = 'none';
	newUser.style.display = 'none';
	newUserButton.style.display = 'none';

	accountsLink.classList.remove("active");
	configLink.classList.remove("active");

	switch (page) {
		case 'config':
			loadConfigPage();
			config.style.display = 'block';
			configLink.classList.add("active");
			break;
		case 'new':
			newUser.style.display = 'block';
			break;
		default:
			loadAccounts();
			accounts.style.display = 'block';
			accountsLink.classList.add("active");
			newUserButton.style.display = 'block';
	}
}

function loadAccountPage() {
	changeAccountPage();
}

function loadAccounts() {
	JSONRPCClient
		.call('AuthService.AllAccounts', {buyer_id: '13672574147039585173'})
		.then(
			(response) => {
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue 
				 */
				if (accountsTable == null) {
					accountsTable = new Vue({
						el: '#accounts',
						data: {
							accounts: response.accounts
						},
						methods: {
							editUser: editUser
						}
					})
				} else {
					Object.assign(accountsTable.$data, {accounts: response.accounts});
				}
			}
		)
		.catch(
			(e) => {
				console.log("Failed to fetch company accounts");
				console.log(e);
			}
		);
}

function loadConfigPage() {
	JSONRPCClient
		.call('BuyersService.GameConfiguration', {buyer_id: '13672574147039585173'})
		.then((response) => {
			userInfo.pubKey = response.game_config.public_key;
			/**
			 * I really dislike this but it is apparently the way to reload/update the data within a vue
			 */
			if (pubKeyInput == null) {
				pubKeyInput = new Vue({
					el: '#pubKey',
					data: {
						pubkey: userInfo.pubKey
					},
					methods: {
						updatePubKey: updatePubKey
					}
				})
			} else {
				Object.assign(pubKeyInput.$data, {pubkey: userInfo.pubKey})
			}
		})
		.catch((e) => {
			console.log("Something went wrong fetching public key");
		});
}

function loadRelayPage() {
	JSONRPCClient
		.call('OpsService.Relays', {})
		.then((response) => {
			/**
			 * I really dislike this but it is apparently the way to reload/update the data within a vue
			 */
			if (relaysTable == null) {
				relaysTable = new Vue({
					el: '#relays',
					data: {
						relays: response.relays || []
					}
				});
			} else {
				Object.assign(relaysTable.$data, {relays: response.relays})
			}
		})
		.catch((e) => {
			console.log("Something went wrong with fetching relays")
		});
}

function loadSessionPage() {
	JSONRPCClient
		.call('BuyersService.Sessions', {buyer_id: '13672574147039585173'})
		.then((response) => {
			if (sessionsTable == null) {
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				sessionsTable = new Vue({
					el: '#sessions',
					data: {
						sessions: response.sessions || []
					},
					methods: {
						fetchSessionInfo: fetchSessionInfo
					}
				});
			} else {
				Object.assign(sessionsTable.$data, {sessions: response.sessions})
			}
		})
		.catch((e) => {
			console.log("Something went wrong with fetching the sessions list");
			console.log(e);
		});
}

function loadUsersPage() {
	// No Endpoint for this yet
}

function updatePubKey() {
	let newPubkey = document.getElementById("pubKey").value;

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {buyer_id: '13672574147039585173', new_public_key: newPubkey})
		.then((response) => {
			document.getElementById("pubKey").value = response.game_config.public_key;
		})
		.catch((e) => {
			console.log("Failed to update public key");
			console.log(e);
		})
}

function fetchSessionInfo(sessionId = '') {

	const id = sessionId || document.getElementById("sessionIDLookup").value;
	document.getElementById("sessionIDLookup").value = '';

	if (id == '') {
		console.log("Can't use a empty id");
		return;
	}
	JSONRPCClient
		.call("BuyersService.Sessions", {buyer_id: '13672574147039585173'/* , session_id: id */})
		.then((response) => {
			var sessionToolMapInstance = new deck.DeckGL({
				mapboxApiAccessToken: mapboxgl.accessToken,
				mapStyle: 'mapbox://styles/mapbox/dark-v10',
				initialViewState: {
					longitude: -98.583333,
					latitude: 39.833333,
					zoom: 4,
					maxZoom: 15,
				},
				controller: true,
				container: 'session-tool-map',
			});

			showDemoChart('latency-chart-1');
			showDemoChart('latency-chart-2');
			showDemoChart('jitter-chart-1');
			showDemoChart('jitter-chart-2');
			showDemoChart('packet-loss-chart-1');
			showDemoChart('packet-loss-chart-2');
			showDemoChart('bandwidth-chart-1');
			showDemoChart('bandwidth-chart-2');
		})
		.catch((e) => {
			console.log("Something went wrong with fetching session information: ");
			console.log(e);
		});
}

function showDemoChart(id) {
	var options = {
		series: [{
			data: [34, 44, 54, 21, 12, 43, 33, 23, 66, 66, 58]
		}],
		chart: {
			type: 'area',
			height: 350,
			toolbar: {
				show: false
			},
			zoom: {
				enabled: false
			},
		},
		legend: {
			show: true
		},
		stroke: {
			curve: 'stepline',
		},
		theme: {
			mode: "light"
		},
		dataLabels: {
			enabled: false
		},
		markers: {
			hover: {
			sizeOffset: 4
			}
		},
		xaxis: {
			lines: {
			show: false,
			}
		},
		yaxis: {
			lines: {
			show: true,
			}
		}
	};

	var chart = new ApexCharts(document.querySelector("#" + id), options);
	chart.render();
}

function editUser(accountInfo) {
	changeAccountPage('new');

	document.getElementById("email").value = accountInfo.email;
	document.getElementById("perms").value = accountInfo.email;
}

JSONRPCClient = {

	async call(method, params) {
		const headers = {
			'Accept':		'application/json',
			'Accept-Encoding':	'gzip',
			'Content-Type':		'application/json',
			'Authorization': `Bearer ${userInfo.token}`
		}

		params = params || {}
		const id = JSON.stringify(params)
		const response = await fetch('/rpc', {
			method: 'POST',
			headers: headers,
			body: JSON.stringify({
            	jsonrpc: '2.0',
				method: method,
				params: params,
				id: id
        	})
		});


		return response.json().then((json) => {
			if (json.error) {
				throw new Error(json.error);
			}
			return json.result
		})
	}
}
window.MapHandler = {
	mapInstance: null,
	async initMap() {
		JSONRPCClient
			.call('BuyersService.SessionsMap', {buyer_id: '13672574147039585173'})
			.then((response) => {
				const DATA_URL =
  					'https://raw.githubusercontent.com/uber-common/deck.gl-data/master/examples/screen-grid/uber-pickup-locations.json';
				const data = DATA_URL, cellSize = 5, gpuAggregation = true, aggregation = 'SUM';
				let sessionGridLayer = new deck.ScreenGridLayer({
					id: 'session-layer',
					data,
					opacity: 0.8,
					getPosition: d => [d[0], d[1]],
					getWeight: d => d[2],
					cellSizePixels: cellSize,
					colorRange: [[0,109,44], [8,81,156]],
					gpuAggregation,
					aggregation
				  })
				var layers = [sessionGridLayer];
				mapInstance = new deck.DeckGL({
					mapboxApiAccessToken: mapboxgl.accessToken,
					mapStyle: 'mapbox://styles/mapbox/dark-v10',
					initialViewState: {
						// Center of the continental US
						longitude: -98.583333,
						latitude: 39.833333,
						zoom: 4,
						// Center of the globe
						/* longitude: 0,
						latitude: 0,
						zoom: 2, */
						maxZoom: 15,
					},
					getColorWeight: (points) => {
						let onNetworkNext = points.find((point) => {
							return point.on_network_next;
						});

						return typeof onNetworkNext === 'undefined' ? 1 : 0;
					},
					container: 'map-workspace',
					controller: true,
					layers: layers,
				});
			})
			.catch((e) => {
				console.log("Something went wrong with map init");
				console.log(e);
			});

		let randomCoord = {
			lat: getRandomInRange(-90, 90, 3),
			lng: getRandomInRange(-180, 180, 3)
		};

		function getRandomInRange(from, to, fixed) {
			return (Math.random() * (to - from) + from).toFixed(fixed) * 1;
		}
	}
}