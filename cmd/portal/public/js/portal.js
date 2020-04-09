/**
 * TODO:
 * 	Refactor all of this into something more reasonable
 */
currentMap = 'maps-vector-equirectangular.png';

function changePage(page) {
	let account = document.getElementById("account-workspace");
	let session = document.getElementById("session-workspace");
	let map = document.getElementById("map-workspace");
	let title = document.getElementById("workspace-title");

	account.style.display = 'none';
	map.style.display = 'none';
	session.style.display = 'none';

	switch (page) {
		case 'sessions':
			session.style.display = 'block';
			title.textContent = 'Session Table';
			break;
		case 'account':
			account.style.display = 'block';
			changeAccountPage();
			title.textContent = 'Account Details';
			break;
		default:
			map.style.display = 'block';
			title.textContent = 'Session Map';
			//reloadMap();
	}
}

function changeAccountPage(page) {
	let config = document.getElementById("config");
	let users = document.getElementById("users");

	config.style.display = 'none';
	users.style.display = 'none';

	switch (page) {
		case 'config':
			config.style.display = 'block';
			break;
		default:
			users.style.display = 'block';
	}
}

function changeMap(map) {
	switch (map) {
		case 'NA':
			currentMap = 'us.png';
			break;
		default:
			currentMap = 'maps-vector-equirectangular.png'
	}
	changePage('home');
}

JSONRPCClient = {

	async call(method, params) {
		const headers = {
			'Accept':		'application/json',
			'Accept-Encoding':	'gzip',
			'Content-Type':		'application/json',
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
		})
	}
}
window.MapHandler = {

	mapInstance: null,
	async initMap() {
		var data = [];

		for (let i = 0; i < 1000000; i++) {
			let lat = getRandomInRange(-90, 90, 3);
			let lon = getRandomInRange(-180, 180, 3);
			data.push({COORDINATES: [lon, lat]});
		}
		mapInstance = new deck.DeckGL({
			mapboxApiAccessToken: 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg',
			mapStyle: 'mapbox://styles/mapbox/dark-v9',
			initialViewState: {
				longitude: -122.45,
				latitude: 37.8,
				zoom: 4
			},
			container: 'map-workspace',
			controller: true,
			layers: [
				new deck.HexagonLayer({
					id: 'hexagon-layer',
					data,
					pickable: false,
					extruded: false,
					radius: 100000,
					elevationScale: 4,
					getPosition: d => d.COORDINATES
				})
			]
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