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

		const query = window.location.search;
		if (query.includes("message=") && query.includes("code=") && query.includes("success=")) {
			let search = query.substring(1);
			let vars = search.split('&');
			let isSignupRedirect = true;
			for (let i = 0; i < vars.length; i++) {
				let pair = vars[i].split('=');
				switch (pair[0]) {
					case "message":
						isSignupRedirect = isSignupRedirect && pair[1] == "Your%20email%20was%20verified.%20You%20can%20continue%20using%20the%20application.";
						break;
					case "code":
						isSignupRedirect = isSignupRedirect && pair[1] == "success";
						break;
					case "success":
						isSignupRedirect = isSignupRedirect && pair[1] == "true";
						break;
				}
			}
			this.isSignupRedirect = isSignupRedirect;
		}

		const isAuthenticated =
			await this.auth0Client.isAuthenticated()
				.catch((e) => {
					Sentry.captureException(e);
				});

		if (isAuthenticated) {
			window.history.replaceState({}, document.title, "/");
			startApp();
			return;
		}
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
	isSignupRedirect: false,
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
	},
	signUp() {
		this.auth0Client.loginWithRedirect({
			connection: "Username-Password-Authentication",
			redirect_uri: window.location.origin,
			screen_hint: "signup"
		}).catch((e) => {
			Sentry.captureException(e);
		});
	},
	resendVerificationEmail() {
		let userId = UserHandler.userInfo.userId;
		let email = UserHandler.userInfo.email;
		JSONRPCClient
			.call("AuthService.ResendVerificationEmail", {
				user_id: userId,
				user_email: email,
				redirect: window.location.origin,
				connection: "Username-Password-Authentication"
			})
			.then((response) => {
				Object.assign(rootComponent.$data.alerts.verifyEmail, {show: false})
				Object.assign(rootComponent.$data.alerts.emailSent, {show: true})
			})
			.catch((error) => {
				console.log("something went wrong with resending verification email")
				Sentry.captureException(error)
				Object.assign(rootComponent.$data.alerts.verifyEmail, {show: false})
				Object.assign(rootComponent.$data.alerts.emailFailed, {show: true})
			})
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
	mapCountLoop: null,
	mapInstance: null,
	sessionToolMapInstance: null,
	initMap() {
		// Not working yet
		// let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateFilter('map', {
			buyerId: "",
			sessionType: 'all'
		});
	},
	updateFilter(filter) {
		Object.assign(rootComponent.$data.pages.map, {filter: filter});
		this.mapCountLoop ? clearInterval(this.mapCountLoop) : null;
		this.mapLoop ? clearInterval(this.mapLoop) : null;

		this.refreshMapCount();
		this.mapCountLoop = setInterval(() => {
			this.refreshMapCount();
		}, 10000);

		this.refreshMapSessions();
		this.mapLoop = setInterval(() => {
			this.refreshMapSessions();
		}, 10000);
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
	refreshMapCount() {
		let filter = rootComponent.$data.pages.map.filter;
		JSONRPCClient
			.call('BuyersService.TotalSessions', {buyer_id: filter.buyerId || ""})
			.then((response) => {
				let direct = response.direct
				let next = response.next

				Object.assign(rootComponent.$data, {
					direct: direct,
					mapSessions: direct + next,
					onNN: next,
				});
			})
			.catch((error) => {
				console.log("Something went wrong fetching map point totals");
				console.log(error);
				Sentry.captureException(error);
			});
	},
	refreshMapSessions() {
		let filter = rootComponent.$data.pages.map.filter;

		JSONRPCClient
			.call('BuyersService.SessionMap', {buyer_id: filter.buyerId || ""})
			.then((response) => {
				let sessions = response.map_points;
				let onNN = sessions.filter((point) => {
					return (point[2] == 1);
				});
				let direct = sessions.filter((point) => {
					return (point[2] == 0);
				});

				const cellSize = 10, aggregation = 'MEAN';
				let gpuAggregation = navigator.appVersion.indexOf("Win") == -1;

				let nnLayer = new deck.ScreenGridLayer({
					id: 'nn-layer',
					data: onNN,
					opacity: 0.8,
					getPosition: d => [d[0], d[1]],
					getWeight: d => 1,
					cellSizePixels: cellSize,
					colorRange: [
						[40, 167, 69],
					],
					gpuAggregation,
					aggregation
				});

				let directLayer = new deck.ScreenGridLayer({
					id: 'direct-layer',
					data: direct,
					opacity: 0.8,
					getPosition: d => [d[0], d[1]],
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
					company: "",
					domain: response.email.split("@")[1],
					email: response.email,
					name: response.name,
					nickname: response.nickname,
					userId: response.sub,
					token: response.__raw,
					verified: response.email_verified,
					roles: []
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

				if (AuthHandler.isSignupRedirect && !UserHandler.isAnonymous() && !UserHandler.isAnonymousPlus() && (!UserHandler.isOwner() || !UserHandler.isAdmin())) {
					JSONRPCClient
						.call("AuthService.UpgradeAccount", {user_id: UserHandler.userInfo.userId})
						.then((response) => {
							let newRoles = response.new_roles || []
							if (newRoles.length > 0) {
								UserHandler.userInfo.roles = response.new_roles;
							}
						})
						.catch((error) => {
							console.log("Something went wrong upgrading the account")
							Sentry.captureException(error)
						})
				}
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
	isAnonymousPlus() {
		return !this.isAnonymous() ? !this.userInfo.verified : false;
	},
	isOwner() {
		return !this.isAnonymous() ? this.userInfo.roles.findIndex((role) => role.name == "Owner") !== -1 : false;
	},
	isViewer() {
		return !this.isAnonymous() ? this.userInfo.roles.findIndex((role) => role.name == "Viewer") !== -1 : false;
	}
}

WorkspaceHandler = {
	mapLoop: null,
	sessionLoop: null,
	sessionToolLoop: null,
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
	changePage(page, options) {
		// Clear all polling loops
		this.sessionLoop ? clearInterval(this.sessionLoop) : null;
		this.mapLoop ? clearInterval(this.mapLoop) : null;
		this.sessionToolLoop ? clearInterval(this.sessionToolLoop) : null;

		switch (page) {
			case 'downloads':
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
				let page =
					UserHandler.isAnonymousPlus() || (
						!UserHandler.isOwner() &&
						!UserHandler.isAnonymous() &&
						!UserHandler.isAdmin()
					) ? "config" : "users";
				Object.assign(rootComponent.$data.pages.settings, {
					newUser: {
						failure: '',
						success: '',
					},
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
				});
				this.changeSettingsPage(page);
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
		if (UserHandler.isAnonymous() || UserHandler.isAnonymousPlus()) {
			return;
		}
		JSONRPCClient
			.call('BuyersService.GameConfiguration', {domain: UserHandler.userInfo.domain})
			.then((response) => {
				UserHandler.userInfo.pubKey = response.game_config.public_key;
				UserHandler.userInfo.company = response.game_config.company;
			})
			.catch((e) => {
				console.log("Something went wrong fetching public key");
				console.log(e)
				Sentry.captureException(e);
				UserHandler.userInfo.pubKey = "";
				UserHandler.userInfo.company = "";
			});

		// Not working / not necessary?
		// let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateAccountsTableFilter({
			buyerId: "",
		});
	},
	loadSessionsPage() {
		// Not working yet
		// let buyerId = !UserHandler.isAdmin() && !UserHandler.isAnonymous() ? UserHandler.userInfo.id : "";
		this.updateSessionFilter({
			buyerId: "",
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

					if (this.sessionToolMapInstance) {
						this.sessionToolMapInstance.setProps({layers: []})
						this.sessionToolMapInstance.setProps({layers: [sessionLocationLayer]})
					} else {
						this.sessionToolMapInstance = new deck.DeckGL({
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
					}
				});
			})
			.catch((e) => {
				if (this.sessionToolLoop) {
					this.changePage('sessions');
					return;
				}
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
			this.sessionToolLoop ? clearInterval(this.sessionToolLoop) : null;
			this.sessionToolLoop = setInterval(() => {
				this.fetchSessionInfo();
			}, 10000);
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

					/**
					 * I really dislike this but it is apparently the way to reload/update the data within a vue
					 */
					Object.assign(rootComponent.$data.pages.sessions, {
						sessions: sessions,
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

	$(document).ready(function() {
		$(window).keydown(function(event){
			if(event.keyCode == 13) {
				event.preventDefault();
				return false;
			}
		});
	});

	UserHandler
		.fetchCurrentUserInfo()
		.then(() => {
			createVueComponents();
			if (UserHandler.isAnonymousPlus()) {
				Object.assign(rootComponent.$data.alerts.verifyEmail, {show: true});
			}
			document.getElementById("app").style.display = 'block';
			WorkspaceHandler.changePage('map');
			JSONRPCClient
				.call('BuyersService.Buyers', {})
				.then((response) => {
					Object.assign(rootComponent.$data, {allBuyers: response.Buyers});
					/* if (UserHandler.isAnonymous()) {
						WorkspaceHandler.welcomeTimeout = setTimeout(() => {
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
							WorkspaceHandler.welcomeTimeout = null;
						}, 30000)
					} */
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
			mapSessions: 0,
			onNN: 0,
			direct: 0,
			alerts: {
				verifyEmail: {
					show: false
				},
				emailSent: {
					show: false
				},
				emailFailed: {
					show: false
				}
			},
			handlers: {
				authHandler: AuthHandler,
				mapHandler: MapHandler,
				userHandler: UserHandler,
				workspaceHandler: WorkspaceHandler,
			},
			modals: {
				signup: {
					email: "",
					show: false,
					showFailure: false,
					showSuccess: false,
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
						bandwidthChart: null,
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
					upgrade: {
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
	let newPubKey = UserHandler.userInfo.pubKey;
	let company = UserHandler.userInfo.company;
	let domain = UserHandler.userInfo.domain

	JSONRPCClient
		.call("BuyersService.UpdateGameConfiguration", {name: company, domain: domain, new_public_key: newPubKey})
		.then((response) => {
			UserHandler.userInfo.pubKey = response.game_config.public_key;
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
	let bandwidthData = [
		[],
		[],
		[],
	];

	data.map((entry) => {
		let timestamp = new Date(entry.timestamp).getTime() / 1000;

		// Latency
		let nextRTT = parseFloat(entry.next.rtt);
		let directRTT = parseFloat(entry.direct.rtt);
		let next = entry.is_multipath && nextRTT >= directRTT ? directRTT : nextRTT;
		let direct = directRTT;
		latencyData.comparison[0].push(timestamp);
		latencyData.comparison[1].push(next);
		latencyData.comparison[2].push(direct);

		// Jitter
		next = parseFloat(entry.next.jitter);
		direct = parseFloat(entry.direct.jitter);
		jitterData.comparison[0].push(timestamp);
		jitterData.comparison[1].push(next);
		jitterData.comparison[2].push(direct);

		// Packetloss
		let nextPL = parseFloat(entry.next.packet_loss);
		let directPL = parseFloat(entry.direct.packet_loss);
		next = entry.is_multipath && nextPL >= directPL ? directPL : nextPL;
		direct = directPL;
		packetLossData.comparison[0].push(timestamp);
		packetLossData.comparison[1].push(next);
		packetLossData.comparison[2].push(direct);

		// Bandwidth
		bandwidthData[0].push(timestamp);
		bandwidthData[1].push(entry.envelope.up);
		bandwidthData[2].push(entry.envelope.down);
	});

	const defaultOpts = {
		width: document.getElementById("latency-chart-1").clientWidth,
		height: 260,
		cursor: {
			drag: {
				x: false,
				y: false
			}
		}
	};

	const latencycomparisonOpts = {
		...defaultOpts,
		scales: {
			"ms": {
				from: "y",
				auto: false,
				range: (self, min, max) => [
					0,
					max,
				],
			}
		},
		series: [
			{
			},
			{
				stroke: "rgb(0, 109, 44)",
				fill: "rgba(0, 109, 44, 0.1)",
				label: "Network Next",
				value: (self, rawValue) => rawValue.toFixed(2)
			},
			{
				stroke: "rgb(49, 130, 189)",
				fill: "rgba(49, 130, 189, 0.1)",
				label: "Direct",
				value: (self, rawValue) => rawValue.toFixed(2)
			},
		],
		axes: [
			{
				show: false
			},
			{
				scale: "ms",
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "ms"),
			}
		],
	};

	const packetLossComparisonOpts = {
		...defaultOpts,
		scales: {
			y: {
				auto: false,
				range: [0, 100],
			}
		},
		series: [
			{},
			{
				stroke: "rgb(0, 109, 44)",
				fill: "rgba(0, 109, 44, 0.1)",
				label: "Network Next",
				value: (self, rawValue) => rawValue.toFixed(2)
			},
			{
				stroke: "rgba(49, 130, 189)",
				fill: "rgba(49, 130, 189, 0.1)",
				label: "Direct",
				value: (self, rawValue) => rawValue.toFixed(2)
			},
		],
		axes: [
			{
				show: false
			},
			{
			  show: true,
			  gap: 5,
			  size: 50,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "%"),
			}
		],
	};

	const bandwidthOpts = {
		...defaultOpts,
		scales: {
			"kbps": {
				from: "y",
				auto: false,
				range: (self, min, max) => [
					0,
					max,
				],
			}
		},
		series: [
			{},
			{
				stroke: "blue",
				fill: "rgba(0,0,255,0.1)",
				label: "Up",
			},
			{
				stroke: "orange",
				fill: "rgba(255,165,0,0.1)",
				label: "Down"
			},
		],
		axes: [
			{
				show: false
			},
			{
				scale: "kbps",
			  show: true,
			  gap: 5,
			  size: 70,
			  values: (self, ticks) => ticks.map(rawValue => rawValue + "kbps"),
			},
			{
				show: false
			}
		]
	};

	if (rootComponent.$data.pages.sessionTool.graphs.latencyComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.latencyComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		latencyComparisonChart: new uPlot(latencycomparisonOpts, latencyData.comparison, document.getElementById("latency-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.jitterComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.jitterComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		jitterComparisonChart: new uPlot(latencycomparisonOpts, jitterData.comparison, document.getElementById("jitter-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.packetLossComparisonChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.packetLossComparisonChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		packetLossComparisonChart: new uPlot(packetLossComparisonOpts, packetLossData.comparison, document.getElementById("packet-loss-chart-1"))
	});

	if (rootComponent.$data.pages.sessionTool.graphs.bandwidthChart != null) {
		rootComponent.$data.pages.sessionTool.graphs.bandwidthChart.destroy();
	}

	Object.assign(rootComponent.$data.pages.sessionTool.graphs, {
		bandwidthChart: new uPlot(bandwidthOpts, bandwidthData, document.getElementById("bandwidth-chart-1"))
	});
}
