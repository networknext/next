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
			title.textContent = 'Account Details';
			break;
		default:
			map.style.display = 'block';
			title.textContent = 'Session Map';
			reloadMap();
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
		JSONRPCClient
			.call('BuyersService.SessionsMap', {buyer_id: '12354645743257'})
			.then((response) => {
				let data = response.on_network_next;
				let nnLayer = new deck.HexagonLayer({
					id: 'nn-layer',
					data,
					colorRange: [
						[49,163,84],
						[49,163,84],
						[49,163,84],
						[49,163,84],
						[49,163,84]
					],
					pickable: true,
					extruded: false,
					radius: 1000,
					elevationScale: 4,
					getPosition: d => d.COORDINATES,
					onHover: info => setTooltip(info.object, info.x, info.y)
				});
				data = response.direct;
				let directLayer = new deck.HexagonLayer({
					id: 'direct-layer',
					data,
					colorRange: [
						[49,130,189],
						[49,130,189],
						[49,130,189],
						[49,130,189],
						[49,130,189],
						[49,130,189],
						[49,130,189]
					],
					pickable: true,
					extruded: false,
					radius: 1000,
					elevationScale: 4,
					getPosition: d => d.COORDINATES,
					onHover: info => setTooltip(info.object, info.x, info.y)
				});
				var layers = [directLayer, nnLayer];
				mapInstance = new deck.DeckGL({
					mapboxApiAccessToken: 'pk.eyJ1IjoiYmF1bWJhY2hhbmRyZXciLCJhIjoiY2s4dDFwcGo2MGowZTNtcXpsbDN6dHBwdyJ9.Sr1lDY9i9o9yz84fJ-PSlg',
					mapStyle: 'mapbox://styles/mapbox/dark-v9',
					initialViewState: {
						// Center of the continental US
						longitude: -98.583333,
						latitude: 39.833333,
						zoom: 4,
						minZoom: 4,
						maxZoom: 15,
					},
					container: 'map-workspace',
					controller: true,
					layers: layers,
					pitch: 80
				});
			})
			.catch((e) => {
				console.log(e);
			});

		function setTooltip(object, x, y) {
			const el = document.getElementById('tooltip');
			if (object) {
				console.log(object)
				el.innerHTML = object.points.length;
				el.style.display = 'block';
				el.style.left = x + 'px';
				el.style.top = y + 'px';
				el.style.fontSize = '32px';
			} else {
				el.style.display = 'none';
			}
		}

		let randomCoord = {
			lat: getRandomInRange(-90, 90, 3),
			lng: getRandomInRange(-180, 180, 3)
		};

		function getRandomInRange(from, to, fixed) {
			return (Math.random() * (to - from) + from).toFixed(fixed) * 1;
		}
	}
}