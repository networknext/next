/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
mapboxgl.accessToken = 'pk.eyJ1Ijoibm5zZWN1cml0eSIsImEiOiJja2FmaXE1Y2cwZGRiMzBub2p3cnE4c3czIn0.3QIueg8fpEy5cBtqRuXMxw';

const DEC_TO_PERC = 100;

var autoSigninPermissions = null;
var addUserPermissions = null;
var editUserPermissions = [];

JSONRPCClient = {
	async call(method, params) {
		let headers = {
			'Accept':		'application/json',
			'Accept-Encoding':	'gzip',
			'Content-Type':		'application/json',
		}
		if (!UserHandler.isAnonymous()) {
				headers['Authorization'] = `Bearer ${UserHandler.userInfo.token}`
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

AuthHandler = {
	async init() {
		const domain = 'networknext.auth0.com';

		// HACK THESE NEED TO BE ENV VARIABLES SOME HOW
		const clientID = window.location.hostname == 'portal.networknext.com' ? 'MaSx99ma3AwYOwWMLm3XWNvQ5WyJWG2Y' : 'oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n';

		this.auth0Client = await createAuth0Client({
			client_id: clientID,
			domain: domain,
		})
		.catch((e) => {
			Sentry.captureException(e);
		});

		const isAuthenticated =
			await this.auth0Client.isAuthenticated()
				.catch((e) => {
					Sentry.captureException(e);
				});

		if (isAuthenticated) {
			startApp();
			return;
		}
		const query = window.location.search;
		if (query.includes("code=") && query.includes("state=")) {

			await this.auth0Client.handleRedirectCallback()
				.catch((e) => {
					Sentry.captureException(e);
				});

			window.history.replaceState({}, document.title, "/");
		}
		startApp();
	},
	auth0Client: null,
	logout() {
		this.auth0Client.logout({ returnTo: window.location.origin });
	},
	login() {
		this.auth0Client.loginWithRedirect({
			connection: "Username-Password-Authentication",
			redirect_uri: window.location.origin
		}).catch((e) => {
			Sentry.captureException(e);
		});
	}
}

MapHandler = {
	defaultUSA: {
		initialViewState: {
			zoom: 4.6,
			longitude: -98.583333, // 'Center' of the US
			latitude: 39.833333,
			minZoom: 5,
			bearing: 0,
			pitch: 0,
		},
	},
	defaultWorld: {
		initialViewState: {
			zoom: 2,
			longitude: 0, // 'Center' of the world map
			latitude: 0,
			minZoom: 2,
			bearing: 0,
			pitch: 0
		},
	},
	mapInstance: null,
	mapLoop: null,
	initMap() {
		let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateFilter('map', {
			buyerId: buyerId,
			sessionType: 'all'
		});

		this.refreshMapSessions();
		this.mapLoop = setInterval(() => {
			this.refreshMapSessions();
		}, 10000);
	},
	updateFilter(filter) {
		Object.assign(rootComponent.$data.pages.map, {filter: filter});
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
				let sessions = response.map_points;
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

				const cellSize = 10, aggregation = 'MEAN';
				let gpuAggregation = navigator.appVersion.indexOf("Win") == -1;

				data = onNN;

				let nnLayer = new deck.ScreenGridLayer({
					id: 'nn-layer',
					data,
					opacity: 0.8,
					getPosition: d => [d.longitude, d.latitude],
					getWeight: d => 1,
					cellSizePixels: cellSize,
					colorRange: [
						[0,109,44],
					],
					gpuAggregation,
					aggregation
				});

				data = direct;

				let directLayer = new deck.ScreenGridLayer({
					id: 'direct-layer',
					data,
					opacity: 0.8,
					getPosition: d => [d.longitude, d.latitude],
					getWeight: d => 1,
					cellSizePixels: cellSize,
					colorRange: [
						[49,130,189],
					],
					gpuAggregation,
					aggregation
				});

				let layers = (onNN.length > 0 || direct.length > 0) ? [directLayer, nnLayer] : [];
				if (this.mapInstance) {
					this.mapInstance.setProps({layers: []})
					this.mapInstance.setProps({layers: layers})
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
	userInfo: null,
	async fetchCurrentUserInfo() {
		return AuthHandler.auth0Client.getIdTokenClaims()
			.then((response) => {
				if (!response) {
					return;
				}
				this.userInfo = {
					email: response.email,
					name: response.name,
					nickname: response.nickname,
					userId: response.sub,
					token: response.__raw,
				};
				return JSONRPCClient.call("AuthService.UserAccount", {user_id: this.userInfo.userId});
			})
			.then((response) => {
				if (!response) {
					return;
				}
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
		return !this.isAnonymous() ? this.userInfo.roles.findIndex((role) => role.name == "Admin") !== -1 : false;
	},
	isAnonymous() {
		return this.userInfo == null;
	},
	isOwner() {
		return !this.isAnonymous() ? this.userInfo.roles.findIndex((role) => role.name == "Owner") !== -1 : false;
	},
	isViewer() {
		return !this.isAnonymous() ? this.userInfo.roles.findIndex((role) => role.name == "Viewer") !== -1 : false;
	},
	signUp() {
		console.log("Signing up")
	}
}

WorkspaceHandler = {
	sessionLoop: null,
	welcomeTimeout: null,
	changeSettingsPage(page) {
		let showSettings = false;
		let showConfig = false;
		switch (page) {
			case 'users':
				showSettings = true;
				break;
			case 'config':
				showConfig = true;
				break;
		}
		Object.assign(rootComponent.$data.pages.settings, {
			showConfig: showConfig,
			showSettings: showSettings,
		});
	},
	changeModal(modal) {
		switch (modal) {
			case 'signup':
				Object.assign(rootComponent.$data.modals.signup, {
					companyName: "",
					email: ""
				});
				break;
		}

		Object.keys(rootComponent.$data.modals).forEach((modal) => {
			Object.assign(rootComponent.$data.modals[modal], {show: false});
		});

		Object.assign(rootComponent.$data.modals[modal], {show: true});

		this.welcomeTimeout !== null ? clearTimeout(this.welcomeTimeout) : null;

		if (!($("#video-modal").data('bs.modal') || {})._isShown) {
			$('#video-modal').modal('toggle');
		}

		$('#video-modal').on('hidden.bs.modal', function () {
			let videoPlayer = document.getElementById("video-player");
			if (videoPlayer) {
				videoPlayer.parentElement.removeChild(videoPlayer)
				videoPlayer.innerHTML = "<div></div>"
			}
		});
	},
	changePage(page, options) {
		// Clear all polling loops
		MapHandler.mapLoop ? clearInterval(MapHandler.mapLoop) : null;
		this.sessionLoop ? clearInterval(this.sessionLoop) : null;

		switch (page) {
			case 'downloads':
				this.loadDownloadsPage();
				break;
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
			case 'settings':
				Object.assign(rootComponent.$data.pages.settings, {
					newUser: {
						failure: '',
						success: '',
					},
					showAccounts: false,
					showConfig: false,
					showSettings: true,
					updateKey: {
						failure: '',
						success: '',
					},
					updateUser: {
						failure: '',
						success: '',
					},
				});
				this.loadSettingsPage();
				break;
			case 'userTool':
				let hash = options || '';
				Object.assign(rootComponent.$data.pages.userTool, {
					danger: false,
					hash: hash,
					info: hash == '',
					showTable: false,
					sessions: []
				});
				hash != '' ? this.fetchUserSessions() : null;
				break;
		}

		Object.keys(rootComponent.$data.pages).forEach((page) => {
			Object.assign(rootComponent.$data.pages[page], {show: false});
		});

		Object.assign(rootComponent.$data.pages[page], {show: true});
	},
	editUser(accountInfo, index) {
		rootComponent.$set(rootComponent.$data.pages.settings.accounts[index], 'delete', false);
		rootComponent.$set(rootComponent.$data.pages.settings.accounts[index], 'edit', true);

		editUserPermissions[accountInfo.user_id].enable();
	},
	saveUser(accountInfo, index) {
		if (accountInfo.edit) {
			let roles = editUserPermissions[accountInfo.user_id].getValue(true);
			JSONRPCClient
				.call('AuthService.UpdateUserRoles', {user_id: `auth0|${accountInfo.user_id}`, roles: roles})
				.then((response) => {
					accountInfo.roles = response.roles;
					this.cancelEditUser(accountInfo);
					Object.assign(rootComponent.$data.pages.settings.updateUser, {
						success: 'Updated user account successfully',
					});
					setTimeout(() => {
						Object.assign(rootComponent.$data.pages.settings.updateUser, {
							success: '',
						});
					}, 5000);
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					Sentry.captureException(e);
					Object.assign(rootComponent.$data.pages.settings.updateUser, {
						failure: 'Failed to update user',
					});
					setTimeout(() => {
						Object.assign(rootComponent.$data.pages.settings.updateUser, {
							failure: '',
						});
					}, 5000);
				});
			return;
		}
		if (accountInfo.delete) {
			JSONRPCClient
				.call('AuthService.DeleteUserAccount', {user_id: `auth0|${accountInfo.user_id}`})
				.then((response) => {
					let accounts = rootComponent.$data.pages.settings.accounts;
					accounts.splice(index, 1);
					Object.assign(rootComponent.$data.pages.settings, {accounts: accounts});
					editUserPermissions[accountInfo.user_id] = null;
					Object.assign(rootComponent.$data.pages.settings.updateUser, {
						success: 'Deleted user account successfully',
					});
					setTimeout(() => {
						Object.assign(rootComponent.$data.pages.settings.updateUser, {
							success: '',
						});
					}, 5000);
				})
				.catch((e) => {
					console.log("Something went wrong updating the users permissions");
					Sentry.captureException(e);
					Object.assign(rootComponent.$data.pages.settings.updateUser, {
						failure: 'Failed to delete user',
					});
					setTimeout(() => {
						Object.assign(rootComponent.$data.pages.settings.updateUser, {
							failure: '',
						});
					}, 5000);
				});
			return;
		}
	},
	deleteUser(index) {
		rootComponent.$set(rootComponent.$data.pages.settings.accounts[index], 'delete', true);
		rootComponent.$set(rootComponent.$data.pages.settings.accounts[index], 'edit', false);
	},
	cancelEditUser(accountInfo, index) {
		editUserPermissions[accountInfo.user_id].disable();
		let accounts = rootComponent.$data.pages.settings.accounts;
		accountInfo.delete = false;
		accountInfo.edit = false;
		accounts[index] = accountInfo;
		Object.assign(rootComponent.$data.pages.settings, {accounts: accounts});
	},
	loadSettingsPage() {
		if (UserHandler.userInfo == null) {
			return;
		}
		if (UserHandler.userInfo.id != '') {
			JSONRPCClient
				.call('BuyersService.GameConfiguration', {buyer_id: UserHandler.userInfo.id})
				.then((response) => {
					UserHandler.userInfo.pubKey = response.game_config.public_key;
				})
				.catch((e) => {
					console.log("Something went wrong fetching public key");
					console.log(e)
					Sentry.captureException(e);
					UserHandler.userInfo.pubKey = "";
				});
		} else {
			UserHandler.userInfo.pubkey = "";
		}

		let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateAccountsTableFilter({
			buyerId: buyerId,
		});
	},
	loadConfigPage() {
		if (UserHandler.isAnonymous()) {
			return;
		}
		JSONRPCClient
			.call('BuyersService.GameConfiguration', {buyer_id: UserHandler.userInfo.id})
			.then((response) => {
				UserHandler.userInfo.pubKey = response.game_config.public_key;
			})
			.catch((e) => {
				console.log("Something went wrong fetching relays");
				Sentry.captureException(e);
			});
		},
		loadDownloadsPage() {
			// Empty for now
		},
	loadRelayPage() {
		JSONRPCClient
			.call('OpsService.Relays', {})
			.then((response) => {
				// Save Relays somewhere
			})
			.catch((e) => {
				console.log("Something went wrong fetching the top sessions list");
				Sentry.captureException(e);
			});
	},
	loadSessionsPage() {
		let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateSessionFilter({
			buyerId: buyerId,
			sessionType: 'all'
		});
		this.refreshSessionTable();
		this.sessionLoop = setInterval(() => {
			this.refreshSessionTable();
		}, 10000);
	},
	fetchSessionInfo() {
		let id = rootComponent.$data.pages.sessionTool.id;

		if (id == '') {
			Object.assign(rootComponent.$data.pages.sessionTool, {
				info: false,
				danger: true,
				meta: null,
				session: null,
				slices: [],
				showDetails: false,
			});
			return;
		}

		JSONRPCClient
			.call("BuyersService.SessionDetails", {session_id: id})
			.then((response) => {
				let meta = response.meta;

				meta.connection = meta.connection == "wifi" ? "Wifi" : meta.connection.charAt(0).toUpperCase() + meta.connection.slice(1);

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

					const NNCOLOR = [0,109,44];
					const DIRECTCOLOR = [49,130,189];

					const cellSize = 10, aggregation = 'MEAN';
					let gpuAggregation = navigator.appVersion.indexOf("Win") == -1;

					let sessionLocationLayer = new deck.ScreenGridLayer({
						id: 'session-location-layer',
						data: [meta],
						opacity: 0.8,
						getPosition: d => [d.location.longitude, d.location.latitude],
						getWeight: d => 1,
						cellSizePixels: cellSize,
						colorRange: meta.on_network_next ? [NNCOLOR] : [DIRECTCOLOR],
						gpuAggregation,
						aggregation
					});

					let sessionToolMapInstance = new deck.DeckGL({
						mapboxApiAccessToken: mapboxgl.accessToken,
						mapStyle: 'mapbox://styles/mapbox/dark-v10',
						initialViewState: {
							zoom: 4,
							longitude: meta.location.longitude, // 'Center' of the world map
							latitude: meta.location.latitude,
							minZoom: 2,
							bearing: 0,
							pitch: 0
						},
						container: 'session-tool-map',
						controller: {
							dragPan: false,
							dragRotate: false
						},
						layers: [sessionLocationLayer],
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
	fetchUserSessions() {
		let hash = rootComponent.$data.pages.userTool.hash;

		if (hash == '') {
			Object.assign(rootComponent.$data.pages.userTool, {
				info: false,
				danger: true,
				sessions: [],
				showTable: false,
			});
			return;
		}

		JSONRPCClient
			.call("BuyersService.UserSessions", {user_hash: hash})
			.then((response) => {
				let sessions = response.sessions;

				Object.assign(rootComponent.$data.pages.userTool, {
					danger: false,
					info: false,
					sessions: sessions,
					showTable: true,
				});
			})
			.catch((e) => {
				Object.assign(rootComponent.$data.pages.userTool, {
					danger: true,
					hash: '',
					info: false,
					sessions: [],
					showTable: false,
				});
				console.log("Something went wrong fetching user sessions: ");
				Sentry.captureException(e);
			});
	},
	loadUsersPage() {
		// No Endpoint for this yet
	},
	updateSessionFilter(filter) {
		Object.assign(rootComponent.$data.pages.sessions, {filter: filter});
	},
	refreshSessionTable() {
		setTimeout(() => {
			let filter = rootComponent.$data.pages.sessions.filter;

			JSONRPCClient
				.call('BuyersService.TopSessions', {buyer_id: filter.buyerId})
				.then((response) => {
					let sessions = response.sessions;
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
	},
	updateAccountsTableFilter(filter) {
		Object.assign(rootComponent.$data.pages.settings, {filter: filter});
		this.refreshAccountsTable();
	},
	refreshAccountsTable() {
		setTimeout(() => {
			let filter = rootComponent.$data.pages.settings.filter;

			let promises = [
				JSONRPCClient
					.call('AuthService.AllAccounts', {}),
				JSONRPCClient
					.call('AuthService.AllRoles', {})
			];
			Promise.all(promises)
				.then((responses) => {
					let accounts = responses[0].accounts;
					allRoles = responses[1].roles;
					console.log(allRoles)

					if (filter.buyerId != '') {
						accounts = accounts.filter((account) => {
							return account.id == filter.buyerId;
						});
					}

					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(rootComponent.$data.pages.settings, {
						accounts: accounts,
						showAccounts: true,
					});

					setTimeout(() => {
						try {
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

							generateRolesDropdown(accounts);
						} catch(e) {
							rootComponent.$data.pages.settings.show ? Sentry.captureException(e) : null;
						}
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
					if (UserHandler.isAnonymous()) {
						WorkspaceHandler.welcomeTimeout = setTimeout(() => {
							WorkspaceHandler.changeModal('welcome');
							WorkspaceHandler.welcomeTimeout = null;
						}, 60000)
					}
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
				authHandler: AuthHandler,
				mapHandler: MapHandler,
				userHandler: UserHandler,
				workspaceHandler: WorkspaceHandler,
			},
			modals: {
				signup: {
					companyName: "",
					email: "",
					show: false,
				},
				welcome: {
					show: false,
				},
			},
			pages: {
				downloads: {
					show: false
				},
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
					graphs: {
						bandwidthUpChart: null,
						bandwidthDownChart: null,
						jitterImprovementChart: null,
						jitterComparisonChart: null,
						latencyImprovementChart: null,
						latencyComparisonChart: null,
						packetLossImprovementChart: null,
						packetLossComparisonChart: null
					}
				},
				settings: {
					accounts: [],
					filter: {
						buyerId: '',
					},
					newUser: {
						failure: '',
						success: '',
					},
					pubKey: '',
					show: false,
					showAccounts: false,
					showConfig: false,
					showSettings: false,
					updateKey: {
						failure: '',
						success: '',
					},
					updateUser: {
						failure: '',
						success: '',
					},
				},
				userTool: {
					danger: false,
					hash: '',
					info: false,
					sessions: [],
					show: false,
					showTable: false,
				}
			}
		},
		methods: {
			addUsers: addUsers,
			saveAutoSignIn: saveAutoSignIn,
			updatePubKey: updatePubKey,
		}
	});
}

function updatePubKey() {
	let newPubkey = document.getElementById("pubkey-input").value;

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {buyer_id: UserHandler.userInfo.id, new_public_key: newPubkey})
		.then((response) => {
			UserHandler.userInfo.pubkey = response.game_config.public_key;
			Object.assign(rootComponent.$data.pages.settings.updateKey, {
				success: 'Updated public key successfully',
			});
			setTimeout(() => {
				Object.assign(rootComponent.$data.pages.settings.updateKey, {
					success: '',
				});
			}, 5000);
		})
		.catch((e) => {
			console.log("Something went wrong updating the public key");
			Sentry.captureException(e);
			Object.assign(rootComponent.$data.pages.settings.updateKey, {
				failure: 'Failed to update public key',
			});
			setTimeout(() => {
				Object.assign(rootComponent.$data.pages.settings.updateKey, {
					failure: '',
				});
			}, 5000);
		});
}

function addUsers() {
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
			Object.assign(rootComponent.$data.pages.settings, {accounts: rootComponent.$data.pages.settings.accounts.concat(newAccounts)});
			setTimeout(() => {
				generateRolesDropdown(newAccounts);
			});
			Object.assign(rootComponent.$data.pages.settings.newUser, {
				success: 'User account added successfully',
			});
			setTimeout(() => {
				Object.assign(rootComponent.$data.pages.settings.newUser, {
					success: '',
				});
			}, 5000);
		})
		.catch((e) => {
			console.log("Something went wrong creating new users");
			Sentry.captureException(e);
			Object.assign(rootComponent.$data.pages.settings.newUser, {
				failure: 'Failed to add user account',
			});
			setTimeout(() => {
				Object.assign(rootComponent.$data.pages.settings.newUser, {
					failure: '',
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

function saveAutoSignIn() {
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
		next = parseFloat(entry.next.packet_loss);
		direct = parseFloat(entry.direct.packet_loss);
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

	if (rootComponent.$data.pages.sessionTool.graphs.latencyImprovementChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.latencyImprovementChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		latencyImprovementChart: new uPlot(latencyImprovementOpts, latencyData.improvement, document.getElementById("latency-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.latencyComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.latencyComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		latencyComparisonChart: new uPlot(latencycomparisonOpts, latencyData.comparison, document.getElementById("latency-chart-2"))
	});

	// let jitterImprovementChart = new uPlot(latencyImprovementOpts, jitterData.improvement, document.getElementById("jitter-chart-1"));
	// let jitterComparisonChart = new uPlot(latencycomparisonOpts, jitterData.comparison, document.getElementById("jitter-chart-2"));

	if (rootComponent.$data.pages.sessionTool.graphs.jitterImprovementChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.jitterImprovementChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		jitterImprovementChart: new uPlot(latencyImprovementOpts, jitterData.improvement, document.getElementById("jitter-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.jitterComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.jitterComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		jitterComparisonChart: new uPlot(latencycomparisonOpts, jitterData.comparison, document.getElementById("jitter-chart-2"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.packetLossImprovementChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.packetLossImprovementChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		packetLossImprovementChart: new uPlot(packetLossImprovementOpts, packetLossData.improvement, document.getElementById("packet-loss-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.packetLossComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.packetLossComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		packetLossComparisonChart: new uPlot(packetLossComparisonOpts, packetLossData.comparison, document.getElementById("packet-loss-chart-2"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.bandwidthUpChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.bandwidthUpChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		bandwidthUpChart: new uPlot(bandwidthUpOpts, bandwidthData.up, document.getElementById("bandwidth-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.bandwidthDownChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.bandwidthDownChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		bandwidthDownChart: new uPlot(bandwidthDownOpts, bandwidthData.down, document.getElementById("bandwidth-chart-2"))
	});
}
