/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg';

const SEC_TO_MS = 1000;
const DEC_TO_PERC = 100;

var userInfo = {
<<<<<<< HEAD
	email: "",
	name: "",
	pubKey: "",
	nickname: "",
	token: "",
	userId: "",
	buyerId: "",
=======
	email: "test",
	name: "test",
	pubKey: "test",
	nickname: "test",
	token: "test",
	userId: "test",
>>>>>>> Turning off auth for dev test
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
	defaultNA: {
		initialViewState: {
			zoom: 4,
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
				let layer = new deck.ScreenGridLayer({
					id: 'sessions-layer',
					data,
					pickable: false,
					opacity: 0.8,
					cellSizePixels: 10,
					colorRange: [
						[0, 25, 0, 25],
						[0, 85, 0, 85],
						[0, 127, 0, 127],
						[0, 170, 0, 170],
						[0, 190, 0, 190],
						[0, 255, 0, 255]
					],
					getPosition: d => [d.longitude, d.latitude],
					getWeight: d => Math.random(10), // Need to come up with a weight system. It won't map anything if the array of points are all identical
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
					container: 'map-workspace',
					controller: true,
					layers: layers,
				});
			})
			.catch((e) => {
				console.log("Something went wrong with map init");
				console.log(e);
			});
	},
	updateMap(mapType) {
		switch (mapType) {
			case 'NA':
				this.mapInstance.setProps({
					...this.defaultNA
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
					console.log("Something went wrong fetching all accounts");
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
	loadSessionPage() {
		JSONRPCClient
			.call('BuyersService.TopSessions', {})
			.then((response) => {
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(sessionsTable.$data, {sessions: response.sessions});
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
	/**
	 * QUESTION: Instead of grabbing the user here can we use the token to then go off and get everything from the backend?
	 * TODO:	 There are 3 different promises going off to get user details. There should be a better way to do this
	 */
	/* Promise.all([
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
	}); */
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

function fetchUserInfo() {
	// Need an endpoint for this
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
			console.log("Something went wrong updating the public key");
			console.log(e);
		})
}

function fetchSessionInfo(sessionId = '') {
	let id = sessionId || document.getElementById("sessionIDLookup").value;
	document.getElementById("sessionIDLookup").value = '';

	if (id == '') {
		console.log("Can't use a empty id");
		return;
	}
	/**
	 * TODO: Add in a catch for when session ID isn't found
	 */
	JSONRPCClient
		.call("BuyersService.SessionDetails", {session_id: id})
		.then((response) => {
			new Vue({
				el: '#sessionDetails',
				data: {
					id: id,
					meta: response.meta,
					slices: response.slices
				},
				methods: {
					fetchSessionInfo: fetchSessionInfo
				}
			});
			let data = {
				latitude: response.meta.location.latitude,
				longitude: response.meta.location.longitude,
			};
			let sessionToolMapInstance = new deck.DeckGL({
				mapboxApiAccessToken: mapboxgl.accessToken,
				mapStyle: 'mapbox://styles/mapbox/dark-v10',
				initialViewState: {
					latitude: data.latitude,
					longitude: data.longitude,
					zoom: 4,
					maxZoom: 15,
				},
				controller: true,
				container: 'session-tool-map',
				/* layers: [
					new deck.IconLayer({
						id: 'icon-layer',
						data,
						pickable: false,
						// iconAtlas and iconMapping are required
						// getIcon: return a string
						iconAtlas: 'marker.png',
						iconMapping: {marker: {x: 0, y: 0, width: 32, height: 32, mask: true}},
						getIcon: d => 'marker',
						sizeScale: 15,
						getPosition: d => [d.longitude, d.latitude],
						getSize: d => 100,
						getColor: d => [7, 140, 0]
					  })
				] */
			});

			generateCharts(response.slices);
		})
		.catch((e) => {
			console.log("Something went wrong fetching session information: ");
			console.log(e);
		});
}


function generateCharts(data) {
	let latencyData = {
		next: [],
		direct: [],
		improvement: [],
	};
	let jitterData = {
		next: [],
		direct: [],
		improvement: [],
	};
	let packetLossData = {
		next: [],
		direct: [],
		improvement: [],
	};
	let bandwidthData = {
		up: [],
		down: [],
	};

	data.map((entry) => {
		let timestamp = new Date(entry.timestamp);
		timestamp = timestamp.toLocaleString(
			'en-us',
			{
				weekday: undefined,
				year: undefined,
				month: undefined,
				day: undefined,
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit',
			},
		);

		// Latency
		let next = Number.parseInt(entry.next.rtt * SEC_TO_MS).toFixed(0);
		let direct = Number.parseInt(entry.direct.rtt * SEC_TO_MS).toFixed(0);
		let improvement = direct - next;
		latencyData.next.push({
			x: timestamp,
			y: next,
		});
		latencyData.direct.push({
			x: timestamp,
			y: direct,
		});
		latencyData.improvement.push({
			x: timestamp,
			y: improvement,
		});

		// Jitter
		next = Number.parseInt(entry.next.rtt * SEC_TO_MS).toFixed(0);
		direct = Number.parseInt(entry.direct.rtt * SEC_TO_MS).toFixed(0);
		improvement = direct - next;
		jitterData.next.push({
			x: timestamp,
			y: next,
		});
		jitterData.direct.push({
			x: timestamp,
			y: direct,
		});
		jitterData.improvement.push({
			x: timestamp,
			y: improvement,
		});

		// Packetloss
		next = Number.parseInt(entry.next.packet_loss * DEC_TO_PERC).toFixed(0);
		direct = Number.parseInt(entry.direct.packet_loss * DEC_TO_PERC).toFixed(0);
		improvement = direct - next;
		packetLossData.next.push({
			x: timestamp,
			y: next,
		});
		packetLossData.direct.push({
			x: timestamp,
			y: direct,
		});
		packetLossData.improvement.push({
			x: timestamp,
			y: improvement,
		});

		// Bandwidth
		bandwidthData.up.push({
			x: timestamp,
			y: entry.envelope.up,
		});
		bandwidthData.down.push({
			x: timestamp,
			y: entry.envelope.down,
		});
	});

	let defaultOptions = {
		chart: {
			type: 'area',
			height: 350,
			toolbar: {
				show: false,
			},
			zoom: {
				enabled: false,
			},
		},
		legend: {
			show: false,
		},
		stroke: {
			curve: 'stepline',
		},
		theme: {
			mode: 'light',
		},
		dataLabels: {
			enabled: false
		},
		markers: {
			hover: {
				sizeOffset: 4,
			},
		},
		xaxis: {
			lines: {
				show: false,
			},
		},
	};

	let latencyOptionsImprovement = {
		series: [
			{
				name: 'Improvement',
				data: latencyData.improvement,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(m/s)',
			},
		},
	};

	let latencyOptionsComparison = {
		series: [
			{
				name: 'Network Next',
				data: latencyData.next,
			},
			{
				name: 'Direct',
				data: latencyData.direct,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(m/s)',
			},
		},
	};

	let latencyImprovementChart = new ApexCharts(
		document.querySelector("#latency-chart-1"),
		{
			...latencyOptionsImprovement,
			...defaultOptions
		},
	);

	let latencyComparisonChart = new ApexCharts(
		document.querySelector("#latency-chart-2"),
		{
			...latencyOptionsComparison,
			...defaultOptions
		},
	);

	let jitterOptionsImprovement = {
		series: [
			{
				name: 'Improvement',
				data: jitterData.improvement,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(m/s)',
			},
		},
	};

	let jitterOptionsComparison = {
		series: [
			{
				name: 'Network Next',
				data: jitterData.next,
			},
			{
				name: 'Direct',
				data: jitterData.direct,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(m/s)',
			},
		},
	};

	let jitterImprovementChart = new ApexCharts(
		document.querySelector("#jitter-chart-1"),
		{
			...jitterOptionsImprovement,
			...defaultOptions
		},
	);

	let jitterComparisonChart = new ApexCharts(
		document.querySelector("#jitter-chart-2"),
		{
			...jitterOptionsComparison,
			...defaultOptions
		},
	);

	let packetLossOptionsImprovement = {
		series: [
			{
				name: 'Improvement',
				data: packetLossData.improvement,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(%)',
			},
		},
	};

	let packetLossOptionsComparison = {
		series: [
			{
				name: 'Network Next',
				data: packetLossData.next,
			},
			{
				name: 'Direct',
				data: packetLossData.direct,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(%)',
			},
		},
	};

	let packetLossImprovementChart = new ApexCharts(
		document.querySelector("#packet-loss-chart-1"),
		{
			...packetLossOptionsImprovement,
			...defaultOptions
		},
	);

	let packetLossComparisonChart = new ApexCharts(
		document.querySelector("#packet-loss-chart-2"),
		{
			...packetLossOptionsComparison,
			...defaultOptions
		},
	);

	let bandwidthOptionsUp = {
		series: [
			{
				name: 'Actual Up',
				data: bandwidthData.up,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(kbps)',
			},
		},
	};

	let bandwidthOptionsDown = {
		series: [
			{
				name: 'Actual Down',
				data: bandwidthData.down,
			},
		],
		yaxis: {
			lines: {
				show: true,
			},
			title: {
				text: '(kbps)',
			},
		},
	};

	let bandwidthUpChart = new ApexCharts(
		document.querySelector("#bandwidth-chart-1"),
		{
			...bandwidthOptionsUp,
			...defaultOptions
		},
	);

	let bandwidthDownChart = new ApexCharts(
		document.querySelector("#bandwidth-chart-2"),
		{
			...bandwidthOptionsDown,
			...defaultOptions
		},
	);
	latencyImprovementChart.render();
	latencyComparisonChart.render();
	jitterImprovementChart.render();
	jitterComparisonChart.render();
	packetLossImprovementChart.render();
	packetLossComparisonChart.render();
	bandwidthUpChart.render();
	bandwidthDownChart.render();
}
