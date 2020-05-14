/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg';

const DEC_TO_PERC = 100;

var defaultSessionDetailsVue = {
	meta: null,
	slices: [],
	showDetails: false,
};

var defaultSessionsTable = {
	sessions: [],
	showCount: false,
	totalSessions: 0,
};

var defaultUserSessionTable = {
	sessions: [],
	showTable: false,
}

var settingsPage = null;
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
			'Authorization': `Bearer ${UserHandler.userInfo.token}`
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
			console.log(json)
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
	initMap() {
		let buyerId = !UserHandler.isAdmin() ? UserHandler.userInfo.id : "";
		this.updateFilter('map', {
			buyerId: buyerId,
			sessionType: 'all'
		});
	},
	updateFilter(filter) {
		Object.assign(rootComponent.$data.pages.map, {filter: filter});
		this.refreshMapSessions();
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
	},
	refreshMapSessions() {
		let filter = rootComponent.$data.pages.map.filter;
		JSONRPCClient
			.call('BuyersService.SessionMapPoints', {buyer_id: filter.buyerId})
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

				let sessions = response.map_points || [];
				let onNN = sessions.filter((point) => {
					return point.on_network_next;
				});
				let direct = sessions.filter((point) => {
					return !point.on_network_next;
				});
				let data = [];

				switch (filter.sessionType) {
					case 'direct':
						data = direct;
						break;
					case 'nn':
						data = onNN;
						break;
					default:
						data = sessions;
				}

				Object.assign(rootComponent.$data, {
					direct: direct,
					mapSessions: sessions,
					onNN: onNN,
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

				let layers = data.length > 0 ? [layer] : [];
				if (this.mapInstance) {
					this.mapInstance.setProps({layers})
				} else {
					this.mapInstance = new deck.DeckGL({
						mapboxApiAccessToken: mapboxgl.accessToken,
						mapStyle: 'mapbox://styles/mapbox/dark-v10',
						initialViewState: {
							...MapHandler.defaultWorld.initialViewState
						},
						container: 'map-container',
						controller: true,
						layers: layers,
					});
				}
				Object.assign(rootComponent.$data, {showCount: true});
			})
			.catch((e) => {
				console.log("Something went wrong with map init");
				console.log(e);
				Sentry.captureException(e);
			});
	},
}

UserHandler = {
	userInfo: {
		company: "",
		email: "",
		id: "",
		name: "",
		nickname: "",
		pubKey: "",
		roles: [],
		token: "",
		userId: "",
	},
	async fetchCurrentUserInfo() {
		return loginClient.getIdTokenClaims()
			.then((response) => {
				this.userInfo = {
					email: response.email,
					name: response.name,
					nickname: response.nickname,
					userId: response.sub,
					token: response.__raw,
				};
				return JSONRPCClient.call("AuthService.UserAccount", {user_id: this.userInfo.userId})
			})
			.then((response) => {
				this.userInfo.id = response.account.id;
				this.userInfo.company = response.account.company_name;
				this.userInfo.roles = response.account.roles;
			}).catch((e) => {
				console.log("Something went wrong getting the current user information");
				console.log(e);
				Sentry.captureException(e);

				// Need to handle no BuyerID gracefully
			});
	},
	isAdmin() {
		return UserHandler.userInfo.roles.findIndex((role) => role.name == "Admin") !== -1
	},
	isOwner() {
		return UserHandler.userInfo.roles.findIndex((role) => role.name == "Owner") !== -1
	},
	isViewer() {
		return UserHandler.userInfo.roles.findIndex((role) => role.name == "Viewer") !== -1
	},
}

WorkspaceHandler = {
	changeAccountPage(page) {
		// Hide all workspaces
		this.pages.accounts.style.display = 'none';
		this.pages.config.style.display = 'none';

		// Run setup for selected page
		switch (page) {
			case 'config':
				this.links.accountsLink.classList.remove("active");
				this.links.configLink.classList.add("active");
				Object.assign(settingsPage.$data, {
					showSettings: false,
				});
				break;
			default:
				this.links.configLink.classList.remove("active");
				this.links.accountsLink.classList.add("active");
				Object.assign(settingsPage.$data, {
					showSettings: true,
				});
		}
	},
	changePage(page, options) {
		switch (page) {
			case 'map':
				MapHandler.initMap();
				break;
			case 'sessions':
				this.loadSessionsPage();
				break;
			case 'sessionTool':
				let id = options || '';
				Object.assign(rootComponent.$data.pages.sessionTool, {
					danger: false,
					id: id,
					info: id == '',
					showDetails: false
				});
				id != '' ? this.fetchSessionInfo() : null;
				break;
		}

		Object.keys(rootComponent.$data.pages).forEach((page) => {
			Object.assign(rootComponent.$data.pages[page], {show: false});
		});

		Object.assign(rootComponent.$data.pages[page], {show: true});
	},
	editUser(accountInfo, index) {
		settingsPage.$set(settingsPage.$data.accounts[index], 'delete', false);
		settingsPage.$set(settingsPage.$data.accounts[index], 'edit', true);

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
					Object.assign(settingsPage.$data.updateUser.success, {
						message: 'Updated user account successfully',
					});
					setTimeout(() => {
						Object.assign(settingsPage.$data.updateUser.success, {
							message: '',
						});
					}, 5000);
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					Sentry.captureException(e);
					Object.assign(settingsPage.$data.updateUser.failure, {
						message: 'Failed to update user',
					});
					setTimeout(() => {
						Object.assign(settingsPage.$data.updateUser.failure, {
							message: '',
						});
					}, 5000);
				});
			return;
		}
		if (accountInfo.delete) {
			JSONRPCClient
				.call('AuthService.DeleteUserAccount', {user_id: `auth0|${accountInfo.user_id}`})
				.then((response) => {
					let accounts = settingsPage.$data.accounts;
					accounts.splice(index, 1);
					Object.assign(settingsPage.$data, {accounts: accounts});
					editUserPermissions[accountInfo.user_id] = null;
					Object.assign(settingsPage.$data.updateUser.success, {
						message: 'Deleted user account successfully',
					});
					setTimeout(() => {
						Object.assign(settingsPage.$data.updateUser.success, {
							message: '',
						});
					}, 5000);
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					Sentry.captureException(e);
					Object.assign(settingsPage.$data.updateUser.failure, {
						message: 'Failed to delete user',
					});
					setTimeout(() => {
						Object.assign(settingsPage.$data.updateUser.failure, {
							message: '',
						});
					}, 5000);
				});
			return;
		}
	},
	deleteUser(index) {
		settingsPage.$set(settingsPage.$data.accounts[index], 'delete', true);
		settingsPage.$set(settingsPage.$data.accounts[index], 'edit', false);
	},
	cancelEditUser(accountInfo, index) {
		editUserPermissions[accountInfo.user_id].disable();
		let accounts = settingsPage.$data.accounts;
		accountInfo.delete = false;
		accountInfo.edit = false;
		accounts[index] = accountInfo;
		Object.assign(settingsPage.$data, {accounts: accounts});
	},
	loadSettingsPage() {
		this.changeAccountPage();

		if (UserHandler.userInfo.id != '') {
			JSONRPCClient
				.call('BuyersService.GameConfiguration', {buyer_id: UserHandler.userInfo.id})
				.then((response) => {
					UserHandler.userInfo.pubKey = response.game_config.public_key;
					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(settingsPage.$data, {
						pubKey: UserHandler.userInfo.pubKey,
					});
				})
				.catch((e) => {
					console.log("Something went wrong fetching public key");
					console.log(e)
					Sentry.captureException(e);
					UserHandler.userInfo.pubKey = "";
					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(settingsPage.$data, {
						pubKey: "",
					});
				});
		} else {
			/**
			 * I really dislike this but it is apparently the way to reload/update the data within a vue
			 */
			Object.assign(settingsPage.$data, {
				pubKey: '',
			});
		}

		let buyerId = !UserHandler.isAdmin() ? UserHandler.userInfo.id : "";
		updateAccountsTableFilter({
			buyerId: buyerId,
		});
	},
	loadConfigPage() {
		JSONRPCClient
			.call('BuyersService.GameConfiguration', {buyer_id: UserHandler.userInfo.id})
			.then((response) => {
				UserHandler.userInfo.pubKey = response.game_config.public_key;
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(pubKeyInput.$data, {
					pubKey: UserHandler.userInfo.pubKey,
				});
			})
			.catch((e) => {
				console.log("Something went wrong fetching relays");
				Sentry.captureException(e);
			});
	},
	loadRelayPage() {
		JSONRPCClient
			.call('OpsService.Relays', {})
			.then((response) => {
				let relays = response.relays;
				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(relaysTable.$data, {relays: relays});
			})
			.catch((e) => {
				console.log("Something went wrong fetching the top sessions list");
				Sentry.captureException(e);
			});
	},
	loadSessionsPage() {
		let buyerId = !UserHandler.isAdmin() ? UserHandler.userInfo.id : "";
		this.updateSessionFilter({
			buyerId: buyerId,
			sessionType: 'all'
		});
	},
	fetchSessionInfo() {
		let id = rootComponent.$data.pages.sessionTool.id;

		if (id == '') {
			Object.assign(rootComponent.$data.pages.sessionTool, {
				info: false,
				danger: true,
			});
			return;
		}

		JSONRPCClient
			.call("BuyersService.SessionDetails", {session_id: id})
			.then((response) => {
				let meta = response.meta;
				meta.nearby_relays = meta.nearby_relays ?? [];
				Object.assign(rootComponent.$data.pages.sessionTool, {
					info: false,
					danger: false,
					meta: meta,
					session: response,
					slices: response.slices,
					showDetails: true,
				});

				setTimeout(() => {
					generateCharts(response.slices);

					var sessionToolMapInstance = new mapboxgl.Map({
						container: 'session-tool-map',
						style: 'mapbox://styles/mapbox/dark-v10',
						center: [meta.location.latitude, meta.location.longitude],
						zoom: 2
					});
				});
			})
			.catch((e) => {
				Object.assign(rootComponent.$data.pages.sessionTool, {
					danger: true,
					id: '',
					info: false,
					meta: null,
					slices: [],
					showDetails: false,
				});
				console.log("Something went wrong fetching session details: ");
				Sentry.captureException(e);
			});
	},
	loadUsersPage() {
		// No Endpoint for this yet
	},
	updateSessionFilter(filter) {
		Object.assign(rootComponent.$data.pages.sessions, {filter: filter});
		this.refreshSessionTable();
	},
	refreshSessionTable() {
		setTimeout(() => {
			let filter = rootComponent.$data.pages.sessions.filter;

			JSONRPCClient
				.call('BuyersService.TopSessions', {buyer_id: filter.buyerId})
				.then((response) => {
					let sessions = response.sessions || [];
					let onNN = sessions.filter((point) => {
						return point.on_network_next;
					});
					let direct = sessions.filter((point) => {
						return !point.on_network_next;
					});

					switch (filter.sessionType) {
						case 'direct':
							data = direct;
							break;
						case 'nn':
							data = onNN;
							break;
						default:
							data = sessions;
					}
					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(rootComponent.$data.pages.sessions, {
						sessions: data,
						showTable: true,
					});
				})
				.catch((e) => {
					console.log("Something went wrong fetching the top sessions list");
					console.log(e);
					Sentry.captureException(e);
				});
		});
	}
}

function startApp() {
	/**
	 * QUESTION: Instead of grabbing the user here can we use the token to then go off and get everything from the backend?
	 * TODO:	 There are 3 different promises going off to get user details. There should be a better way to do this
	 */

	UserHandler
		.fetchCurrentUserInfo()
		.then(() => {
			createVueComponents();
			document.getElementById("app").style.display = 'block';
			WorkspaceHandler.changePage('map');
			JSONRPCClient
				.call('BuyersService.Buyers', {})
				.then((response) => {
					Object.assign(rootComponent.$data, {allBuyers: response.Buyers});
				})
				.catch((e) => {
					console.log("Something went wrong initializing the map");
					console.log(e);
					Sentry.captureException(e);
				});
		}).catch((e) => {
			console.log("Something went wrong getting the current user information");
			console.log(e);
			Sentry.captureException(e);
		});
}

function createVueComponents() {
	rootComponent = new Vue({
		el: '#root',
		data: {
			allBuyers: [],
			showCount: false,
			mapSessions: [],
			onNN: [],
			direct: [],
			handlers: {
				mapHandler: MapHandler,
				userHandler: UserHandler,
				workspaceHandler: WorkspaceHandler,
			},
			pages: {
				map: {
					filter: {
						buyerId: '',
						sessionType: '',
					},
					show: false,
				},
				sessions: {
					filter: {
						buyerId: '',
						sessionType: '',
					},
					sessions: [],
					show: false,
					showTable: false,
				},
				sessionTool: {
					danger: false,
					id: '',
					info: false,
					meta: null,
					session: null,
					show: false,
					showDetails: false,
					showFailure: false,
					showSuccess: false,
					slices: [],
				},
				settings: {
					accounts: [],
					pubKey: '',
					show: false,
					showAccounts: true,
					showConfig: false,
				},
				userTool: {
					failure: {
						message: ''
					},
					sessions: [],
					show: false,
					showFailure: false,
					showSessions: false,
					showSuccess: false,
					showTable: false,
					success: {
						message: ''
					},
				}
			}
		},
		methods: {
		}
	});

	/* settingsPage = new Vue({
		el: '#settings-page',
		data: {
			accounts: null,
			allBuyers: allBuyers,
			filter: {buyerId: UserHandler.userInfo.id},
			newUser: {
				failure: {
					message: '',
				},
				success: {
					message: '',
				},
			},
			pubKey: UserHandler.userInfo.pubKey,
			showAccounts: false,
			showSettings: true,
			updateUser: {
				failure: {
					message: '',
				},
				success: {
					message: '',
				},
			},
		},
		methods: {
			cancelEditUser: WorkspaceHandler.cancelEditUser,
			deleteUser: WorkspaceHandler.deleteUser,
			editUser: WorkspaceHandler.editUser,
			isAdmin: UserHandler.isAdmin,
			isOwner: UserHandler.isOwner,
			refreshAccountsTable: refreshAccountsTable,
			saveUser: WorkspaceHandler.saveUser,
			updateAccountsTableFilter: updateAccountsTableFilter,
			updatePubKey: updatePubKey
		}
	});
	mapSessionsCount = new Vue({
		el: '#map-sessions-count',
		data: {
			allBuyers: allBuyers,
			filter: {buyerId: UserHandler.userInfo.id, sessionType: 'all'},
			showCount: false,
			sessions: [],
			userInfo: UserHandler.userInfo,
			onNN: [],
		},
		methods: {
			isAdmin: UserHandler.isAdmin,
			refreshMapSessions: MapHandler.refreshMapSessions,
			updateFilter: MapHandler.updateFilter,
		}
	});
	sessionDetailsVue = new Vue({
		el: '#session-details',
		data: {
			meta: null,
			slices: [],
			showDetails: false,
			allBuyers: allBuyers
		}
	});
	sessionsTable = new Vue({
		el: '#sessions',
		data: {
			allBuyers: allBuyers,
			filter: {buyerId: UserHandler.userInfo.id, sessionType: 'all'},
			...defaultSessionsTable,
			allMapSessions: [],
			nnSessions: [],
		},
		methods: {
			fetchSessionInfo: fetchSessionInfo,
			fetchUserSessions: fetchUserSessions,
			isAdmin: UserHandler.isAdmin,
			updateSessionFilter: updateSessionFilter,
			refreshSessionTable: refreshSessionTable,
		}
	});
	userSessionTable = new Vue({
		el: '#user-sessions',
		data: {...defaultUserSessionTable},
		methods: {
			fetchSessionInfo: fetchSessionInfo
		}
	}); */
}

function updateAccountsTableFilter(filter) {
	Object.assign(settingsPage.$data, {filter: filter});
	refreshAccountsTable();
}

function refreshAccountsTable() {
	setTimeout(() => {
		let filter = settingsPage.$data.filter;

		let promises = [
			JSONRPCClient
				.call('AuthService.AllAccounts', {buyer_id: filter.buyerId}),
			JSONRPCClient
				.call('AuthService.AllRoles', {})
		];
		Promise.all(promises)
			.then((responses) => {
				allRoles = responses[1].roles;
				let accounts = responses[0].accounts || [];

				/**
				 * I really dislike this but it is apparently the way to reload/update the data within a vue
				 */
				Object.assign(settingsPage.$data, {
					accounts: accounts,
					showAccounts: true,
				});

				setTimeout(() => {
					let choices = allRoles.map((role) => {
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

					choices = allRoles.map((role) => {
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

					choices = allRoles.map((role) => {
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

					choices = allRoles.map((role) => {
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

					generateRolesDropdown(accounts);
				});
			}
		)
		.catch((errors) => {
			console.log("Something went wrong loading settings page");
			console.log(errors);
			Sentry.captureException(errors);
		});
	});
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
			Sentry.captureException(e);
			WorkspaceHandler.alerts.userToolDanger.style.display = 'block';
			document.getElementById("user-hash-input").value = '';
		});
}

function updatePubKey() {
	let newPubkey = document.getElementById("pubkey-input").value;

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {buyer_id: UserHandler.userInfo.id, new_public_key: newPubkey})
		.then((response) => {
			UserHandler.userInfo.pubkey = response.game_config.public_key;
		})
		.catch((e) => {
			console.log("Something went wrong updating the public key");
			Sentry.captureException(e);
		});
}

function addUsers(event) {
	event.preventDefault();
	let roles = addUserPermissions.getValue(true);
	let emails = String(document.getElementById("new-user-emails").value)
		.split(/(,|\n)/g)
		.map((x) => x.trim())
		.filter((x) => x !== "" && x !== ",");

	if (roles.length == 0) {
		roles = [{
			description: "Can see current sessions and the map.",
			id: "rol_ScQpWhLvmTKRlqLU",
			name: "Viewer",
		}];
	}
	JSONRPCClient
		.call("AuthService.AddUserAccount", {emails: emails, roles: roles})
		.then((response) => {
			let newAccounts = response.accounts;
			Object.assign(settingsPage.$data, {accounts: settingsPage.$data.accounts.concat(newAccounts)});
			setTimeout(() => {
				generateRolesDropdown(newAccounts);
			});
			Object.assign(settingsPage.$data.newUser.success, {
				message: 'User account added successfully',
			});
			setTimeout(() => {
				Object.assign(settingsPage.$data.newUser.success, {
					message: '',
				});
			}, 5000);
		})
		.catch((e) => {
			console.log("Something went wrong creating new users");
			Sentry.captureException(e);
			Object.assign(settingsPage.$data.newUser.failure, {
				message: 'Failed to add user account',
			});
			setTimeout(() => {
				Object.assign(settingsPage.$data.newUser.failure, {
					message: '',
				});
			}, 5000);
		});
	addUserPermissions.removeActiveItems();
	document.getElementById("new-user-emails").value = '';
}

function generateRolesDropdown(accounts) {
	accounts.forEach((account) => {
		if (!editUserPermissions[account.user_id]) {
			editUserPermissions[account.user_id] = new Choices(
				document.getElementById(`edit-user-permissions-${account.user_id}`),
				{
					removeItemButton: true,
					choices: allRoles.map((role) => {
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
}

function saveAutoSignIn(event) {
	event.preventDefault();
	let roles = autoSigninPermissions.getValue(true);
	let domain = document.getElementById("auto-sign-in-domain").value;

	// Make JSONRPC call
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