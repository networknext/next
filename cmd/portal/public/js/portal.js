/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg';

const DEC_TO_PERC = 100;

var userInfo = {
	email: "",
	name: "",
	pubKey: "",
	nickname: "",
	token: "",
	userId: "",
};

var defaultSessionDetailsVue = {
	meta: null,
	slices: [],
	showDetails: false,
};

var defaultSessionsTable = {
	onNN: [],
	sessions: [],
	showCount: false,
};

var defaultUserSessionTable = {
	sessions: [],
	showTable: false,
}

var accountsTable = null;
var mapSessionsCount = null;
var pubKeyInput = null;
var relaysTable = null;
var sessionDetailsVue = null;
var sessionsTable = null;
var userSessionTable = null;

var autoSigninPermissions = null;
var addUserPermissions = null;
var editUserPermissions = [];

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
		});
	}
}

MapHandler = {
	defaultUSA: {
		initialViewState: {
			zoom: 4.6,
			longitude: -98.583333, // 'Center' of the US
			latitude: 39.833333,
			maxZoom: 14,
		},
	},
	defaultWorld: {
		initialViewState: {
			zoom: 2,
			longitude: 0, // 'Center' of the world map
			latitude: 0,
			maxZoom: 14,
		},
	},
	mapInstance: null,
	async initMap() {
		JSONRPCClient
			.call('BuyersService.SessionMapPoints', {})
			.then((response) => {
				/**
				 * This code is used for demo purposes -> it uses around 580k points over NYC
				 */
				/* const DATA_URL =
					  'https://raw.githubusercontent.com/uber-common/deck.gl-data/master/examples/screen-grid/uber-pickup-locations.json';
				let data = DATA_URL;
				const cellSize = 5, gpuAggregation = true, aggregation = 'SUM';
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
				}); */
				let data = response.map_points;
				Object.assign(mapSessionsCount.$data, {
					onNN: data.filter((point) => {
						return point.on_network_next == true;
					}),
					sessions: data,
				});
				let layer = new deck.ScreenGridLayer({
					id: 'sessions-layer',
					data,
					pickable: false,
					opacity: 0.8,
					cellSizePixels: 10,
					colorRange: [
						[40, 167, 69],
						[36, 163, 113],
						[27, 153, 159],
						[18, 143, 206],
						[9, 133, 252],
						[0, 123, 255]
					],
					getPosition: d => [d.longitude, d.latitude],
					getWeight: d => d.on_network_next ? 100 : 1, // Need to come up with a weight system. It won't map anything if the array of points are all identical
					gpuAggregation: true,
					aggregation: 'SUM'
				});
				let layers = [layer];
				this.mapInstance = new deck.DeckGL({
					mapboxApiAccessToken: mapboxgl.accessToken,
					mapStyle: 'mapbox://styles/mapbox/dark-v10',
					initialViewState: {
						...this.defaultWorld.initialViewState
					},
					container: 'map-container',
					controller: true,
					layers: layers,
				});
				Object.assign(mapSessionsCount.$data, {showCount: true});
			})
			.catch((e) => {
				console.log("Something went wrong with map init");
				console.log(e);
			});
	},
	updateMap(mapType) {
		switch (mapType) {
			case 'USA':
				this.mapInstance.setProps({
					...this.defaultUSA
				});
				break;
			case 'WORLD':
				this.mapInstance.setProps({
					...this.defaultWorld
				});
				break;
			default:
				// Nothing for now
		}
	}
}

WorkspaceHandler = {
	alerts: {
		sessionToolAlert: document.getElementById("session-tool-alert"),
		sessionToolDanger: document.getElementById("session-tool-danger"),
		userToolAlert: document.getElementById("user-tool-alert"),
		userToolDanger: document.getElementById("user-tool-danger"),
	},
	links: {
		accountsLink: document.getElementById("accounts-link"),
		configLink: document.getElementById("config-link"),
		mapLink: document.getElementById("map-link"),
		sessionsLink: document.getElementById("sessions-link"),
		sessionToolLink: document.getElementById("session-tool-link"),
		settingsLink: document.getElementById("settings-link"),
		userToolLink: document.getElementById("user-tool-link"),
	},
	newUserEmail: document.getElementById("email"),
	newUserPerms: document.getElementById("perms"),
	pages: {
		accounts: document.getElementById("accounts-page"),
		config: document.getElementById("config-page"),
	},
	spinners: {
		map: document.getElementById("map-spinner"),
		sessions: document.getElementById("sessions-spinner"),
		sessionTool: document.getElementById("session-tool-spinner"),
		userTool: document.getElementById("user-tool-spinner"),
		settings: document.getElementById("settings-spinner"),
	},
	workspaces: {
		mapWorkspace: document.getElementById("map-workspace"),
		sessionsWorkspace: document.getElementById("sessions-workspace"),
		sessionToolWorkspace: document.getElementById("session-tool-workspace"),
		settingsWorkspace: document.getElementById("settings-workspace"),
		userToolWorkspace: document.getElementById("user-tool-workspace"),
	},
	changeAccountPage(page) {
		// Hide all workspaces
		this.pages.accounts.style.display = 'none';
		this.pages.config.style.display = 'none';

		// Run setup for selected page
		switch (page) {
			case 'config':
				this.loadConfigPage();
				this.links.accountsLink.classList.remove("active");
				this.links.configLink.classList.add("active");
				this.pages.config.style.display = 'block';
				break;
			default:
				this.links.configLink.classList.remove("active");
				this.links.accountsLink.classList.add("active");
				this.pages.accounts.style.display = 'block';
		}
	},
	changePage(page) {
		// Hide all workspaces
		this.workspaces.mapWorkspace.style.display = 'none';
		this.workspaces.sessionsWorkspace.style.display = 'none';
		this.workspaces.sessionToolWorkspace.style.display = 'none';
		this.workspaces.settingsWorkspace.style.display = 'none';
		this.workspaces.userToolWorkspace.style.display = 'none';

		// Remove all link highlights
		this.links.accountsLink.classList.remove("active");
		this.links.configLink.classList.remove("active");
		this.links.mapLink.classList.remove("active");
		this.links.sessionsLink.classList.remove("active");
		this.links.sessionToolLink.classList.remove("active");
		this.links.userToolLink.classList.remove("active");
		this.links.settingsLink.classList.remove("active");

		// Run setup for selected page
		switch (page) {
			case 'settings':
				this.loadSettingsPage();
				this.workspaces.settingsWorkspace.style.display = 'block';
				this.links.accountsLink.classList.add("active");
				this.links.settingsLink.classList.add("active");
				break;
			case 'sessions':
				this.loadSessionsPage();
				this.workspaces.sessionsWorkspace.style.display = 'block';
				this.links.sessionsLink.classList.add("active");
				break;
			case 'session-tool':
				WorkspaceHandler.alerts.sessionToolAlert.style.display = 'block';
				WorkspaceHandler.alerts.sessionToolDanger.style.display = 'none';

				Object.assign(sessionDetailsVue.$data, {...defaultSessionDetailsVue});
				document.getElementById("session-id-input").value = '';
				this.workspaces.sessionToolWorkspace.style.display = 'block';
				this.links.sessionToolLink.classList.add("active");
				break;
			case 'user-tool':
				WorkspaceHandler.alerts.userToolAlert.style.display = 'block';
				WorkspaceHandler.alerts.userToolDanger.style.display = 'none';

				Object.assign(userSessionTable.$data, {...defaultUserSessionTable});

				document.getElementById("user-hash-input").value = '';
				this.workspaces.userToolWorkspace.style.display = 'block';
				this.links.userToolLink.classList.add("active");
				break;
			default:
				this.workspaces.mapWorkspace.style.display = 'block';
				this.links.mapLink.classList.add("active");
		}
	},
	editUser(accountInfo, index) {
		accountsTable.$set(accountsTable.$data.accounts[index], 'delete', false);
		accountsTable.$set(accountsTable.$data.accounts[index], 'edit', true);

		editUserPermissions[accountInfo.user_id].enable();
	},
	saveUser(accountInfo, index) {
		if (accountInfo.edit) {
			let roles = editUserPermissions[accountInfo.user_id].getValue(true);
			JSONRPCClient
				.call('AuthService.UpdateUserRoles', {user_id: `auth0|${accountInfo.user_id}`, roles: roles})
				.then((response) => {
					accountInfo.roles = response.roles || [];
					WorkspaceHandler.cancelEditUser(accountInfo);
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					console.log(e);
				});
			return;
		}
		if (accountInfo.delete) {
			JSONRPCClient
				.call('AuthService.DeleteUserAccount', {user_id: `auth0|${accountInfo.user_id}`})
				.then((response) => {
					let accounts = accountsTable.$data.accounts;
					accounts.splice(index, 1);
					Object.assign(accountsTable.$data, {accounts: accounts});
					editUserPermissions[accountInfo.user_id] = null;
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					console.log(e);
				});
			return;
		}
	},
	deleteUser(index) {
		accountsTable.$set(accountsTable.$data.accounts[index], 'delete', true);
		accountsTable.$set(accountsTable.$data.accounts[index], 'edit', false);
	},
	cancelEditUser(accountInfo, index) {
		editUserPermissions[accountInfo.user_id].disable();
		let accounts = accountsTable.$data.accounts;
		accountInfo.delete = false;
		accountInfo.edit = false;
		accounts[index] = accountInfo;
		Object.assign(accountsTable.$data, {accounts: accounts});
	},
	loadSettingsPage() {
		this.changeAccountPage();
		let promises = [
			JSONRPCClient
				.call('AuthService.AllAccounts', {buyer_id: '13672574147039585173'}),
			JSONRPCClient
				.call('AuthService.AllRoles', {})
		];
		Promise.all(promises)
			.then(
				(responses) => {
					let roles = responses[1].roles;
					let accounts = responses[0].accounts;
					let choices = roles.map((role) => {
						return {
							value: role,
							label: role.name,
							customProperties: {
								description: role.description,
							},
						};
					});

					if (!addUserPermissions) {
						addUserPermissions = new Choices(
							document.getElementById("add-user-permissions"),
							{
								removeItemButton: true,
								choices: choices,
							}
						);
					}

					choices = roles.map((role) => {
						return {
							value: role,
							label: role.name,
							customProperties: {
								description: role.description,
							},
							selected: role.name === 'Viewer'
						};
					});

					if (!autoSigninPermissions) {
						autoSigninPermissions = new Choices(
							document.getElementById("auto-signin-permissions"),
							{
								removeItemButton: true,
								choices: choices,
							}
						);
					}

					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(accountsTable.$data, {
						accounts: accounts,
						showAccounts: true,
					});

					setTimeout(() => {
						accounts.forEach((account) => {
							if (!editUserPermissions[account.user_id]) {
								editUserPermissions[account.user_id] = new Choices(
									document.getElementById(`edit-user-permissions-${account.user_id}`),
									{
										removeItemButton: true,
										choices: roles.map((role) => {
											return {
												value: role,
												label: role.name,
												customProperties: {
													description: role.description,
												},
												selected: account.roles.findIndex((userRole) => role.name == userRole.name) !== -1
											};
										}),
									}
								).disable();
							}
						});
					});
				}
			)
			.catch((errors) => {
				console.log("Something went wrong loading settings page");
				console.log(errors);
			});
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
				console.log(e);
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
				console.log("Something went wrong fetching relays");
				console.log(e);
			});
	},
	loadSessionsPage() {
		JSONRPCClient
			.call('BuyersService.TopSessions', {})
			.then((response) => {
				let sessions = response.sessions;
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(sessionsTable.$data, {
					onNN: sessions.filter((session) => {
						return session.on_network_next;
					}),
					sessions: sessions,
					showCount: true,
				});
			})
			.catch((e) => {
				console.log("Something went wrong fetching the top sessions list");
				console.log(e);
			});
	},
	loadUsersPage() {
		// No Endpoint for this yet
	}
}

function startApp() {
	/**
	 * QUESTION: Instead of grabbing the user here can we use the token to then go off and get everything from the backend?
	 * TODO:	 There are 3 different promises going off to get user details. There should be a better way to do this
	 */
	Promise.all([
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
		createVueComponents();
		document.getElementById("app").style.display = 'block';
		MapHandler
			.initMap()
			.then((response) => {
				console.log("Map init successful");
			})
			.catch((e) => {
				console.log("Something went wrong initializing the map");
				console.log(e);
			});
	}).catch((e) => {
		console.log("Something went wrong getting the current user information");
	});
}

function createVueComponents() {
	accountsTable = new Vue({
		el: '#accounts',
		data: {
			accounts: null,
			showAccounts: false,
		},
		methods: {
			cancelEditUser: WorkspaceHandler.cancelEditUser,
			deleteUser: WorkspaceHandler.deleteUser,
			editUser: WorkspaceHandler.editUser,
			saveUser: WorkspaceHandler.saveUser,
		}
	});
	mapSessionsCount = new Vue({
		el: '#map-sessions-count',
		data: {
			onNN: [],
			sessions: [],
			showCount: false,
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
	sessionDetailsVue = new Vue({
		el: '#session-details',
		data: {
			meta: null,
			slices: [],
			showDetails: false,
		}
	});
	sessionsTable = new Vue({
		el: '#sessions',
		data: {...defaultSessionsTable},
		methods: {
			fetchSessionInfo: fetchSessionInfo,
			fetchUserSessions: fetchUserSessions,
		}
	});
	userSessionTable = new Vue({
		el: '#user-sessions',
		data: {...defaultUserSessionTable},
		methods: {
			fetchSessionInfo: fetchSessionInfo
		}
	})
}

function fetchUserSessions(userHash = '') {
	WorkspaceHandler.alerts.userToolAlert.style.display = 'none';
	WorkspaceHandler.alerts.userToolDanger.style.display = 'none';

	let hash = '';

	if (userHash != '') {
		WorkspaceHandler.changePage('user-tool');
		hash = userHash;
		document.getElementById("user-hash-input").value = hash;
	} else {
		hash = document.getElementById("user-hash-input").value;
	}

	if (hash == '') {
		Object.assign(userSessionTable.$data, {...defaultUserSessionTable});
		document.getElementById("user-hash-input").value = '';
		WorkspaceHandler.alerts.userToolDanger.style.display = 'block';
		return;
	}

	JSONRPCClient
		.call("BuyersService.UserSessions", {user_hash: hash})
		.then((response) => {
			let sessions = response.sessions || [];

			/**
			 * I really dislike this but it is apparently the way to reload/update the data within a vue
			 */
			Object.assign(userSessionTable.$data, {
				sessions: sessions,
				showTable: true,
			});

			WorkspaceHandler.alerts.userToolAlert.style.display = 'none';
			WorkspaceHandler.alerts.userToolDanger.style.display = 'none';
		})
		.catch((e) => {
			Object.assign(userSessionTable.$data, {...defaultUserSessionTable});
			console.log("Something went wrong fetching user sessions: ");
			console.log(e);
			WorkspaceHandler.alerts.userToolDanger.style.display = 'block';
			document.getElementById("user-hash-input").value = '';
		});
}

function updatePubKey() {
	let newPubkey = document.getElementById("pubkey-input").value;

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {buyer_id: '13672574147039585173', new_public_key: newPubkey})
		.then((response) => {
			userInfo.pubkey = response.game_config.public_key;
		})
		.catch((e) => {
			console.log("Something went wrong updating the public key");
			console.log(e);
		});
}

function fetchSessionInfo(sessionId = '') {
	WorkspaceHandler.alerts.sessionToolAlert.style.display = 'none';
	WorkspaceHandler.alerts.sessionToolDanger.style.display = 'none';

	let id = '';

	if (sessionId != '') {
		WorkspaceHandler.changePage('session-tool');
		id = sessionId;
		document.getElementById("session-id-input").value = id;
	} else {
		id = document.getElementById("session-id-input").value;
	}

	if (id == '') {
		Object.assign(sessionDetailsVue.$data, {...defaultSessionDetailsVue});
		document.getElementById("session-id-input").value = '';
		WorkspaceHandler.alerts.sessionToolDanger.style.display = 'block';
		return;
	}

	JSONRPCClient
		.call("BuyersService.SessionDetails", {session_id: id})
		.then((response) => {
			let meta = response.meta;
			meta.nearby_relays = meta.nearby_relays ?? [];
			Object.assign(sessionDetailsVue.$data, {
				meta: meta,
				slices: response.slices,
				showDetails: true,
			});

			setTimeout(() => {
				generateCharts(response.slices);

				var sessionToolMapInstance = new mapboxgl.Map({
					container: 'session-tool-map',
					style: 'mapbox://styles/mapbox/dark-v10',
					center: [0, 0],
					zoom: 2
				});
			});

			WorkspaceHandler.alerts.sessionToolAlert.style.display = 'none';
			WorkspaceHandler.alerts.sessionToolDanger.style.display = 'none';
		})
		.catch((e) => {
			Object.assign(sessionDetailsVue.$data, {...defaultSessionDetailsVue});
			console.log("Something went wrong fetching session details: ");
			console.log(e);
			WorkspaceHandler.alerts.sessionToolDanger.style.display = 'block';
			document.getElementById("session-id-input").value = '';
		});
}


function generateCharts(data) {
	let latencyData = {
		improvement: [
			[],
			[],
		],
		comparison: [
			[],
			[],
			[],
		],
	};
	let jitterData = {
		improvement: [
			[],
			[],
		],
		comparison: [
			[],
			[],
			[],
		],
	};
	let packetLossData = {
		improvement: [
			[],
			[],
		],
		comparison: [
			[],
			[],
			[],
		],
	};
	let bandwidthData = {
		up: [
			[],
			[],
		],
		down: [
			[],
			[],
		],
	};

	data.map((entry) => {
		let timestamp = new Date(entry.timestamp).getTime() / 1000;

		// Latency
		let next = parseFloat(entry.next.rtt);
		let direct = parseFloat(entry.direct.rtt);
		let improvement = direct - next;
		latencyData.improvement[0].push(timestamp);
		latencyData.improvement[1].push(improvement);
		latencyData.comparison[0].push(timestamp);
		latencyData.comparison[1].push(next);
		latencyData.comparison[2].push(direct);

		// Jitter
		next = parseFloat(entry.next.jitter);
		direct = parseFloat(entry.direct.jitter);
		improvement = direct - next;
		jitterData.improvement[0].push(timestamp);
		jitterData.improvement[1].push(improvement);
		jitterData.comparison[0].push(timestamp);
		jitterData.comparison[1].push(next);
		jitterData.comparison[2].push(direct);

		// Packetloss
		next = parseFloat(entry.next.packet_loss * DEC_TO_PERC);
		direct = parseFloat(entry.direct.packet_loss * DEC_TO_PERC);
		improvement = direct - next;
		packetLossData.improvement[0].push(timestamp);
		packetLossData.improvement[1].push(improvement);
		packetLossData.comparison[0].push(timestamp);
		packetLossData.comparison[1].push(next);
		packetLossData.comparison[2].push(direct);

		// Bandwidth
		bandwidthData.up[0].push(timestamp);
		bandwidthData.up[1].push(entry.envelope.up);
		bandwidthData.down[0].push(timestamp);
		bandwidthData.down[1].push(entry.envelope.down);
	});

	const defaultOpts = {
		width: document.getElementById("latency-chart-1").clientWidth,
		height: 260,
	};

	const latencyImprovementOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "green",
				fill: "rgba(0,255,0,0.1)",
				label: "Improvement",
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "ms"),
			}
		  ]
	};

	const latencycomparisonOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "blue",
				fill: "rgba(0,0,255,0.1)",
				label: "Network Next",
			},
			{
				stroke: "red",
				fill: "rgba(255,0,0,0.1)",
				label: "Direct",
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "ms"),
			}
		  ]
	};

	const packetLossImprovementOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "green",
				fill: "rgba(0,255,0,0.1)",
				label: "Improvement",
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 50,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "%"),
			}
		  ]
	};

	const packetLossComparisonOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "blue",
				fill: "rgba(0,0,255,0.1)",
				label: "Network Next",
			},
			{
				stroke: "red",
				fill: "rgba(255,0,0,0.1)",
				label: "Direct",
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 50,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "%"),
			}
		]
	};

	const bandwidthUpOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "blue",
				fill: "rgba(0,0,255,0.1)",
				label: "Actual Up",
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "kbps"),
			}
		]
	};

	const bandwidthDownOpts = {
		...defaultOpts,
		series: [
			{},
			{
				stroke: "orange",
				fill: "rgba(255,165,0,0.1)",
				label: "Actual Down"
			},
		],
		axes: [
			{},
			{
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "kbps"),
			}
		]
	};

	let latencyImprovementChart = new uPlot(latencyImprovementOpts, latencyData.improvement, document.getElementById("latency-chart-1"));
	let latencyComparisonChart = new uPlot(latencycomparisonOpts, latencyData.comparison, document.getElementById("latency-chart-2"));

	let jitterImprovementChart = new uPlot(latencyImprovementOpts, jitterData.improvement, document.getElementById("jitter-chart-1"));
	let jitterComparisonChart = new uPlot(latencycomparisonOpts, jitterData.comparison, document.getElementById("jitter-chart-2"));

	let packetLossImprovementChart = new uPlot(packetLossImprovementOpts, packetLossData.improvement, document.getElementById("packet-loss-chart-1"));
	let packetLossComparisonChart = new uPlot(packetLossComparisonOpts, packetLossData.comparison, document.getElementById("packet-loss-chart-2"));

	let bandwidthUpChart = new uPlot(bandwidthUpOpts, bandwidthData.up, document.getElementById("bandwidth-chart-1"));
	let bandwidthDownChart = new uPlot(bandwidthDownOpts, bandwidthData.down, document.getElementById("bandwidth-chart-2"));
}