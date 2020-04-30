/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg';

var userInfo = {
	email: "",
	name: "",
	pubKey: "",
	nickname: "",
	token: "",
	userId: "",
};

var accountsTable = null;
var pubKeyInput = null;
var relaysTable = null;
var sessionsTable = null;

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

MapHandler = {
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
	}
}

WorkspaceHandler = {
	accountWorkspacePages: {
		configPage: document.getElementById("config"),
		newUserPage: document.getElementById("new-user"),
	},
	links: {
		accountsLink: document.getElementById("accounts-link"),
		configLink: document.getElementById("config-link"),
		mapLink: document.getElementById("home-link"),
		relaysLink: document.getElementById("relays-link"),
		sessionsLink: document.getElementById("sessions-link"),
		usersLink: document.getElementById("users-link"),
	},
	showAccountsTable: false,
	workspaceTitle: document.getElementById("workspace-title"),
	newUserEmail: document.getElementById("email"),
	newUserPerms: document.getElementById("perms"),
	workspaces: {
		accountsWorkspace: document.getElementById("accounts-workspace"),
		mapWorkspace: document.getElementById("map-workspace"),
		relaysWorkspace: document.getElementById("relays-workspace"),
		sessionsWorkspace: document.getElementById("sessions-workspace"),
		usersWorkspace: document.getElementById("users-workspace"),
	},
	changeAccountPage(page) {
		let newUserButton = document.getElementById("new-user-button");

		//Hide all workspace pages
		this.accountWorkspacePages.configPage.style.display = 'none';
		this.accountWorkspacePages.newUserPage.style.display = 'none';

		//Hide the accounts table Vue
		Object.assign(accountsTable.$data, {showAccountsTable: false});

		//Hide the new user button
		newUserButton.style.display = 'none';

		//Remove all link highlights
		this.links.accountsLink.classList.remove("active");
		this.links.configLink.classList.remove("active");

		//Run setup for selected account page
		switch (page) {
			case 'config':
				this.loadConfigPage();
				this.accountWorkspacePages.configPage.style.display = 'block';
				this.links.configLink.classList.add("active");
				break;
			case 'new':
				this.accountWorkspacePages.newUserPage.style.display = 'block';
				this.newUserEmail.value = '';
				this.newUserPerms.value = '';
				break;
			default:
				this.loadAccounts();
				Object.assign(accountsTable.$data, {showAccountsTable: true});
				this.links.accountsLink.classList.add("active");
				newUserButton.style.display = 'block';
		}
	},
	changePage(page) {
		// Hide all workspaces
		this.workspaces.accountsWorkspace.style.display = 'none';
		this.workspaces.mapWorkspace.style.display = 'none';
		this.workspaces.relaysWorkspace.style.display = 'none';
		this.workspaces.sessionsWorkspace.style.display = 'none';
		this.workspaces.usersWorkspace.style.display = 'none';

		// Remove all link highlights
		this.links.mapLink.classList.remove("active");
		this.links.relaysLink.classList.remove("active");
		this.links.sessionsLink.classList.remove("active");
		this.links.usersLink.classList.remove("active");

		// Run setup for selected page
		switch (page) {
			case 'account':
				this.changeAccountPage();
				this.workspaces.accountsWorkspace.style.display = 'block';
				this.workspaceTitle.textContent = 'Account Details';
				break;
			case 'relay':
				this.loadRelayPage();
				this.workspaces.relaysWorkspace.style.display = 'block';
				this.links.relaysLink.classList.add("active");
				this.workspaceTitle.textContent = 'Relays Table';
				break;
			case 'sessions':
				this.loadSessionPage();
				this.workspaces.sessionsWorkspace.style.display = 'block';
				this.links.sessionsLink.classList.add("active");
				this.workspaceTitle.textContent = 'Session Table';
				break;
			case 'users':
				this.loadUsersPage();
				this.workspaces.usersWorkspace.style.display = 'block';
				this.links.usersLink.classList.add("active");
				this.workspaceTitle.textContent = 'User Table';
				break;
			default:
				this.workspaces.mapWorkspace.style.display = 'block';
				this.links.mapLink.classList.add("active");
				this.workspaceTitle.textContent = 'Session Map';
		}
	},
	editUser(accountInfo) {
		WorkspaceHandler.changeAccountPage('new');

		WorkspaceHandler.newUserEmail.value = accountInfo.email || '';
		WorkspaceHandler.newUserPerms.value = accountInfo.email || '';
	},
	loadAccounts() {
		JSONRPCClient
			.call('AuthService.AllAccounts', {buyer_id: '13672574147039585173'})
			.then(
				(response) => {
					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(accountsTable.$data, {accounts: response.accounts});
				}
			)
			.catch(
				(e) => {
					console.log("Failed to fetch company accounts");
					console.log(e);
				}
			);
	},
	loadConfigPage() {
		JSONRPCClient
			.call('BuyersService.GameConfiguration', {buyer_id: '13672574147039585173'})
			.then((response) => {
				userInfo.pubKey = response.game_config.public_key;
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(pubKeyInput.$data, {pubKey: userInfo.pubKey});
			})
			.catch((e) => {
				console.log("Something went wrong fetching public key");
			});
	},
	loadRelayPage() {
		JSONRPCClient
			.call('OpsService.Relays', {})
			.then((response) => {
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(relaysTable.$data, {relays: response.relays});
			})
			.catch((e) => {
				console.log("Something went wrong with fetching relays")
			});
	},
	loadSessionPage() {
		JSONRPCClient
			.call('BuyersService.Sessions', {buyer_id: '13672574147039585173'})
			.then((response) => {
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(sessionsTable.$data, {sessions: response.sessions});
			})
			.catch((e) => {
				console.log("Something went wrong with fetching the sessions list");
				console.log(e);
			});
	},
	loadUsersPage() {
		// No Endpoint for this yet
	}
}

function startApp() {
	createVueComponents();
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

function createVueComponents() {
	accountsTable = new Vue({
		el: '#accounts',
		data: {
			accounts: null,
			showAccountsTable: false
		},
		methods: {
			editUser: WorkspaceHandler.editUser
		}
	});
	pubKeyInput = new Vue({
		el: '#pubKey',
		data: {
			pubKey: userInfo.pubKey
		},
		methods: {
			updatePubKey: updatePubKey
		}
	});
	relaysTable = new Vue({
		el: '#relays',
		data: {
			relays: null
		}
	});
	sessionsTable = new Vue({
		el: '#sessions',
		data: {
			sessions: null
		},
		methods: {
			fetchSessionInfo: fetchSessionInfo
		}
	});
	usersTable = new Vue({
		el: '#users',
		data: {
			users: null
		},
		methods: {
			fetchUserInfo: fetchUserInfo
		}
	});
}

function updatePubKey() {
	let newPubkey = document.getElementById("pubKey").value;

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {buyer_id: '13672574147039585173', new_public_key: newPubkey})
		.then((response) => {
			userInfo.pubkey = response.game_config.public_key;
			document.getElementById("pubKey").value = userInfo.pubKey;
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

function fetchUserInfo(userID) {
	// Need an endpoint for this
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

	var chart = new ApexCharts(document.querySelector(`#${id}`), options);
	chart.render();
}
