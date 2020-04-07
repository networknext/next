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
		image.src = currentMap;

		// Image setup function for first load
		function setupMapImage() {
			// Get the height and width of the canvas container
			canvas.width = this.naturalWidth;
			canvas.height = this.naturalHeight;

			// Draw the image to screen
			ctx.drawImage(this, 0, 0, this.width, this.height);

			let points = [];

			let hexbin = d3.hexbin()
				.size([this.width, this.height])
				.radius(10); // TODO: Change this hexsize to be configurable

			let color = d3.scale.linear()
				.domain([14, 15, 35, 132])
				.range(["#333", "#d7191c", "#ffffbf", "#2c7bb6"])
				.interpolate(d3.interpolateHcl);

			imageData = ctx.getImageData(0, 0, this.width, this.height).data;

			// Rescale the colors.
			for (var c, i = 0, n = this.width * this.height * 4, d = imageData; i < n; i += 4) {
				points.push([i/4%this.width, Math.floor(i/4/this.width), d[i]]);
			}

			var grid_odd_r = makeHexOddR("#grid-offset-odd-r", 10, this.width, this.height);

			let mapper = {};
			hexagons = hexbin(points);
			hexagons.forEach(function(d) {
				d.mean = d3.mean(d, function(p) { return p[2]; });
				d.x = d3.mean(d, function(p) { return p[0]; });
				d.y = d3.mean(d, function(p) { return p[1]; });

				var s = new ScreenCoordinate(d.x, d.y);
				s.scale(grid_odd_r.grid.scale / 2);
				d.cube = FractionalCube.cubeRound(grid_odd_r.grid.cartesianToHex(s));
				mapper[d.cube.toString()] = d.mean;
			});


			grid_odd_r.tiles
				.each(function(d) {
					d.center = grid_odd_r.grid.hexToCenter(d.cube);
					d.color = color(mapper[d.key]);
					d.node.select("polygon")
					.style("fill", d.color);
				});

			grid_odd_r.hexgrid = hexWorldGrid(grid_odd_r);
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

		function screenToLonLat(grid_odd_r, screenPoint) {
			var lon_intp = d3.interpolate(-180.0, 180.0);
			var lat_intp = d3.interpolate(-90.0, 90.0);
			var lon = lon_intp(screenPoint[0] / grid_odd_r.width);
			var lat = lat_intp(screenPoint[1] / grid_odd_r.height);
			// since screen coordinate is starting at the top-left corner
			// we need to invert the Y-axis, i.e. lat
			return [lon, -lat];
		}

		function hexToLonLatHex(grid_odd_r, hex) {
			var hexvertices = grid_odd_r.grid.polygonVertices();
			var center = grid_odd_r.grid.hexToCenter(hex);
			var hexpoints = [];
			hexvertices.forEach(function(v,i) {
				var sx = center.x + v.x;
				var sy = center.y + v.y;
				var lonlat = screenToLonLat(grid_odd_r, [sx, sy]);
				hexpoints.push(lonlat);
			});
			hexpoints.push(hexpoints[0]); // close the polygon
			hexpoints.reverse(); // polygon needs to arrange in counter-clockwise

			return hexpoints;
		  }

		function makeHexOddR(svg, hexsize, width, height) {
			var hex_width = hexsize * 2 * Math.sin(Math.PI / 3);
			var hex_height = hexsize * 1.5;

			d3.selectAll("svg > *").remove();

			var grid_odd_r =  makeGridDiagram(d3.select(svg),
								Grid.trapezoidalShape(0, (width/hex_width)-1, 0, height/hex_height, Grid.oddRToCube))
								  .addHexCoordinates(Grid.cubeToOddR, true, false)
								  .update(hexsize*2, true);
			grid_odd_r.width = width;
			grid_odd_r.height = height;
			return grid_odd_r;
		}

		function hexWorldGrid(grid_odd_r) {
			var hgrid = [];
			grid_odd_r.tiles
			  	.each(function(d) {
					if (d.center !== undefined) {
						hexpoints = hexToLonLatHex(grid_odd_r, d.cube);
						var feature = {
							type: "Feature",
							geometry: {
								type: "LineString",
								coordinates: hexpoints
							},
							properties: {
								"color": d.color
							}
						};
						hgrid.push(feature);
					}
				});
			console.log("HexGrid size is " + hgrid.length);
			return {
				type: "FeatureCollection",
				features: hgrid
			};
		}

		var idle_tracker = {
			interval: 1000,
			idle_threshold: 1000,
			running: false,
			needs_to_run: false,
			last_activity: Date.now(),
			callback: null
		};
		idle_tracker.user_activity = function(e) {
			this.last_activity = Date.now();
		}
		idle_tracker.loop = function() {
			idle_tracker.running = setTimeout(idle_tracker.loop, idle_tracker.interval);
			if (idle_tracker.needs_to_run || Date.now() - idle_tracker.last_activity > idle_tracker.idle_threshold) {
				idle_tracker.callback();
			}
			idle_tracker.needs_to_run = false;
		}
		idle_tracker.start = function() {
			this.needs_to_run = true;
			if (!this.running) {
				// There is no loop running so start it, and also start tracking user idle
				this.running = setTimeout(this.loop, 0);
				window.addEventListener('scroll', this.user_activity);
				window.addEventListener('touchmove', this.user_activity);
			} else {
				// There's a loop scheduled but I want it to run immediately
				clearTimeout(this.running);
				this.running = setTimeout(this.loop, 1);
			}
		};
		idle_tracker.stop = function() {
			if (this.running) {
				// Stop tracking user idle when we don't need to (performance)
				window.removeEventListener('scroll', this.user_activity);
				window.removeEventListener('touchmove', this.user_activity);
				clearTimeout(this.running);
				this.running = false;
			}
		}

		delay.queue = d3.map();  // which elements need redrawing?
		delay.refresh = d3.set();  // set of elements we've seen before
		idle_tracker.callback = _delayDrawOnTimeout;
		window.addEventListener('scroll', _delayedDraw);
		window.addEventListener('resize', _delayedDraw);

		function makeGridDiagram(svg, cubes) {
			var diagram = {};

			diagram.nodes = cubes.map(function(n) { return {cube: n, key: n.toString()}; });
			diagram.root = svg.append('g');
			diagram.tiles = diagram.root.selectAll("g.tile").data(diagram.nodes, function(node) { return node.key; });
			diagram.tiles.enter()
				.append('g').attr('class', "tile")
				.each(function(d) { d.node = d3.select(this); });
			diagram.polygons = diagram.tiles.append('polygon');


			diagram.makeTilesSelectable = function(callback) {
				diagram.selected = d3.set();
				diagram.toggle = function(cube) {
					if (diagram.selected.has(cube)) {
						diagram.selected.remove(cube);
					} else {
						diagram.selected.add(cube);
					}
				};

				var drag_state = 0;
				var drag = d3.behavior.drag()
					.on('dragstart', function(d) {
						drag_state = diagram.selected.has(d.cube);
					})
					.on('drag', function() {
						var target = d3.select(d3.event.sourceEvent.target);
						if (target !== undefined && target.data()[0] && target.data()[0].cube) {
							var cube = target.data()[0].cube;
							if (drag_state) {
								diagram.selected.remove(cube);
							} else {
								diagram.selected.add(cube);
							}
						}
						callback();
					});

				diagram.tiles
					.on('click', function(d) {
						d3.event.preventDefault();
						diagram.toggle(d.cube);
						callback();
					})
					.call(drag);
			};


			diagram.addLabels = function(labelFunction) {
				diagram.tiles.append('text')
					.attr('y', "0.4em")
					.text(function(d, i) { return labelFunction? labelFunction(d, i) : ""; });
				return diagram;
			};


			diagram.addHexCoordinates = function(converter, withMouseover, withText) {
				diagram.nodes.forEach(function (n) { n.hex = converter(n.cube); });
				if (withText) {
				diagram.tiles.append('text')
					.attr('y', "0.4em")
					.each(function(d) {
						var selection = d3.select(this);
						selection.append('tspan').attr('class', "q").text(d.hex.q);
						selection.append('tspan').text(", ");
						selection.append('tspan').attr('class', "r").text(d.hex.r);
					});
				}

				function setSelection(hex) {
					diagram.tiles
						.classed('q-axis-same', function(other) { return hex.q == other.hex.q; })
						.classed('r-axis-same', function(other) { return hex.r == other.hex.r; });
				}

				if (withMouseover) {
					diagram.tiles
						.on('mouseover', function(d) {
							setSelection(d.hex);
						})
						.on('touchstart', function(d) {
							setSelection(d.hex);
						});
				}

				return diagram;
			};

			diagram.addCubeCoordinates = function(withMouseover) {
				diagram.tiles.append('text')
					.each(function(d) {
						var selection = d3.select(this);
						var labels = [d.cube.x, d.cube.y, d.cube.z];
						if (labels[0] == 0 && labels[1] == 0 && labels[2] == 0) {
							// Special case: label the origin with x/y/z so that you can tell where things to
							labels = ["x", "y", "z"];
						}
						selection.append('tspan').attr('class', "q").text(labels[0]);
						selection.append('tspan').attr('class', "s").text(labels[1]);
						selection.append('tspan').attr('class', "r").text(labels[2]);
					});

				function relocate() {
					var BL = 4;  // adjust to vertically center
					var offsets = diagram.orientation? [14, -9+BL, -14, -9+BL, 0, 13+BL] : [13, 0+BL, -9, -14+BL, -9, 14+BL];
					offsets = offsets.map(function(f) { return f * diagram.scale / 50; });
					diagram.tiles.select(".q").attr('x', offsets[0]).attr('y', offsets[1]);
					diagram.tiles.select(".s").attr('x', offsets[2]).attr('y', offsets[3]);
					diagram.tiles.select(".r").attr('x', offsets[4]).attr('y', offsets[5]);
				}

				function setSelection(cube) {
					["q", "s", "r"].forEach(function (axis, i) {
						diagram.tiles.classed(axis + "-axis-same", function(other) { return cube.v()[i] == other.cube.v()[i]; });
					});
				}

				if (withMouseover) {
					diagram.tiles
						.on('mouseover', function(d) { return setSelection(d.cube); })
						.on('touchstart', function(d) { return setSelection(d.cube); });
				}

				diagram.onUpdate(relocate);
				return diagram;
			};


			diagram.addPath = function() {
				diagram.pathLayer = this.root.append('path')
					.attr('d', "M 0 0")
					.attr('class', "path");
				diagram.setPath = function(path) {
					var d = [];
					for (var i = 0; i < path.length; i++) {
						d.push(i == 0? 'M' : 'L');
						d.push(diagram.grid.hexToCenter(path[i]));
					}
					diagram.pathLayer.attr('d', d.join(" "));
				};
			};


			var pre_callbacks = [];
			var post_callbacks = [];
			diagram.onLayout = function(callback) { pre_callbacks.push(callback); };
			diagram.onUpdate = function(callback) { post_callbacks.push(callback); };

			var hexagon_points = makeHexagonShape(diagram.scale);

			diagram.update = function(scale, orientation) {
				if (scale != diagram.scale) {
					diagram.scale = scale;
					hexagon_points = makeHexagonShape(scale);
					diagram.polygons.attr('points', hexagon_points);
				}
				diagram.orientation = orientation;

				pre_callbacks.forEach(function (f) { f(); });
				var grid = new Grid(scale, orientation, diagram.nodes.map(function(node) { return node.cube; }));
				var bounds = grid.bounds();
				var first_draw = !diagram.grid;
				diagram.grid = grid;

				delay(svg, function(animate) {
					if (first_draw) { animate = function(selection) { return selection; }; }

					// NOTE: In Webkit I can use svg.node().clientWidth but in Gecko that returns 0 :(
					diagram.translate = new ScreenCoordinate((parseFloat(svg.attr('width')) - bounds.minX - bounds.maxX)/2,
															(parseFloat(svg.attr('height')) - bounds.minY - bounds.maxY)/2);
					animate(diagram.root)
						.attr('transform', "translate(" + diagram.translate + ")");

					animate(diagram.tiles)
						.attr('transform', function(node) {
							var center = grid.hexToCenter(node.cube);
							return "translate(" + center.x + "," + center.y + ")";
						});

					animate(diagram.polygons)
						.attr('transform', "rotate(" + (orientation * -30) + ")");

					post_callbacks.forEach(function (f) { f(); });
				});

				return diagram;
			};

			return diagram;
		}

		function makeHexagonShape(scale) {
			return hexToPolygon(scale, 0, 0, false).map(function(p) { return p.x.toFixed(3) + "," + p.y.toFixed(3); }).join(" ");
		}

		function hexToPolygon(scale, x, y, orientation) {
			// NOTE: the article says to use angles 0..300 or 30..330 (e.g. I
			// add 30 degrees for pointy top) but I instead use -30..270
			// (e.g. I subtract 30 degrees for pointy top) because it better
			// matches the animations I needed for my diagrams. They're
			// equivalent.
			var points = [];
			for (var i = 0; i < 6; i++) {
				var angle = 2 * Math.PI * (2*i - orientation) / 12;
				points.push(new ScreenCoordinate(x + 0.5 * scale * Math.cos(angle),
												 y + 0.5 * scale * Math.sin(angle)));
			}
			return points;
		}

		function distanceToScreen(node) {
			// Compare two ranges: the top:bottom of the browser window and
			// the top:bottom of the element
			var viewTop = window.pageYOffset;
			var viewBottom = viewTop + window.innerHeight;

			// Workaround for Firefox: SVG nodes have no
			// offsetTop/offsetHeight so I check the parent instead
			if (node.offsetTop === undefined) {
				node = node.parentNode;
			}
			var elementTop = node.offsetTop;
			var elementBottom = elementTop + node.offsetHeight;
			return Math.max(0, elementTop - viewBottom, viewTop - elementBottom);
		}

		function delay(element, action) {
			delay.queue.set(element.attr('id'), [element, action]);
			idle_tracker.start();
		}

		function _delayedDraw() {
			var actions = [];
			var idle_draws_allowed = 4;

			// First evaluate all the actions and how far the elements are from being viewed
			delay.queue.forEach(function(id, ea) {
				var element = ea[0], action = ea[1];
				var d = distanceToScreen(element.node());
				actions.push([id, action, d]);
			});

			// Sort so that the ones closest to the viewport are first
			actions.sort(function(a, b) { return a[2] - b[2]; });

			// Draw all the ones that are visible now, or up to
			// idle_draws_allowed that aren't visible now
			actions.forEach(function(ia) {
				var id = ia[0], action = ia[1], d = ia[2];
				if (d == 0 || idle_draws_allowed > 0) {
					if (d != 0) --idle_draws_allowed;

					delay.queue.remove(id);
					delay.refresh.add(id);

					var animate = delay.refresh.has(id) && d == 0;
					action(function(selection) {
						return animate? selection.transition().duration(200) : selection;
					});
				}
			});
		}

		function _delayDrawOnTimeout() {
			_delayedDraw();
			if (delay.queue.keys().length == 0) { idle_tracker.stop(); }
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