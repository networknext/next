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
		})

		return response.json().then((json) => {
			if (json.error) {
				throw new Error(json.error)
			}
			return json.result
		})
	}
}
window.MapHandler = {

	async initMap() {
		console.log("Initializing map");
		// Grab the canvas element
		document.getElementById('canvas');

		// Setup the drawing context
		const ctx = canvas.getContext('2d');

		// Init a new image object
		const image = new Image();
		// Tell JS what to do when the image first loads
		image.onload = setupMapImage;

		// Give the object the image you want to use
		image.src = './maps-vector-equirectangular.png';

		// Image setup function for first load
		function setupMapImage() {
			// Get the height and width of the canvas container
			canvas.width = this.naturalWidth;
			canvas.height = this.naturalHeight;

			// Draw the image to screen
			ctx.drawImage(this, 0, 0, this.width, this.height);

			// Get the correct offset to use going forward for plotting points
			let offset = {
				x: (canvas.width / 360),
				y: (canvas.height / 180)
			};

			// Translate the coordinate given to you to something that is mappable to the canvas
			let coord = translateCoord({lat: 42.654110, lng: -73.752650}, offset);
			// Draw the pixel at that location
			drawPixelAtLocation(coord);

			// Make a bunch of random points for testing
			for (var i = 0; i < 1000; i++) {
				let coord = translateCoord(
				{
					lat: getRandomInRange(-90, 90, 3),
					lng: getRandomInRange(-180, 180, 3)
				}, offset)

				drawPixelAtLocation(coord);
			}

		}
		function drawPixelAtLocation(coord) {
			ctx.rect(coord.x, coord.y, 1, 1);
			ctx.fillStyle = 'red';
			ctx.fill()
		}

		function getRandomInRange(from, to, fixed) {
			return (Math.random() * (to - from) + from).toFixed(fixed) * 1;
		}

		function translateCoord(coord, offset) {
			return {
				x: (coord.lng + 180) * offset.x,
				y: (90 - coord.lat) * offset.y
			}
		}
	}
}

function reloadMap() {
	MapHandler
		.initMap()
		.then(() => {
			console.log("Map Initialized");
		})
		.catch((e) => {
			console.log("Something went wrong with the map init!");
			console.log(e);
		});
}