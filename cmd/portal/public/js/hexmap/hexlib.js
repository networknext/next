// hexlib code is following Amit's Hex Grid implementation at
// https://www.redblobgames.com/grids/hexagons/implementation.html
//
(function (console, $hx_exports) { "use strict";
var Cube = $hx_exports.Cube = function(x,y,z) {
	this.x = x;
	this.y = y;
	this.z = z;
};
Cube.add = function(a,b) {
	return new Cube(a.x + b.x,a.y + b.y,a.z + b.z);
};
Cube.scale = function(a,k) {
	return new Cube(a.x * k,a.y * k,a.z * k);
};
Cube.direction = function(direction) {
	return Cube.directions[direction];
};
Cube.neighbor = function(hex,direction) {
	return Cube.add(hex,Cube.direction(direction));
};
Cube.diagonalNeighbor = function(hex,direction) {
	return Cube.add(hex,Cube.diagonals[direction]);
};
Cube.distance = function(a,b) {
	return Std["int"]((Math.abs(a.x - b.x) + Math.abs(a.y - b.y) + Math.abs(a.z - b.z)) / 2);
};
Cube.$length = function(h) {
	return Std["int"]((Math.abs(h.x) + Math.abs(h.y) + Math.abs(h.z)) / 2);
};
Cube.prototype = {
	toString: function() {
		return this.v().join(",");
	}
	,v: function() {
		return [this.x,this.y,this.z];
	}
	,rotateLeft: function() {
		return new Cube(-this.y,-this.z,-this.x);
	}
	,rotateRight: function() {
		return new Cube(-this.z,-this.x,-this.y);
	}
	,equals: function(other) {
		return this.x == other.x && this.y == other.y && this.z == other.z;
	}
};
var FractionalCube = $hx_exports.FractionalCube = function(x,y,z) {
	this.x = x;
	this.y = y;
	this.z = z;
};
FractionalCube.cubeRound = function(h) {
	var rx = Math.round(h.x);
	var ry = Math.round(h.y);
	var rz = Math.round(h.z);
	var x_diff = Math.abs(rx - h.x);
	var y_diff = Math.abs(ry - h.y);
	var z_diff = Math.abs(rz - h.z);
	if(x_diff > y_diff && x_diff > z_diff) rx = -ry - rz; else if(y_diff > z_diff) ry = -rx - rz; else rz = -rx - ry;
	return new Cube(rx,ry,rz);
};
FractionalCube.cubeLerp = function(a,b,t) {
	return new FractionalCube(a.x + (b.x - a.x) * t,a.y + (b.y - a.y) * t,a.z + (b.z - a.z) * t);
};
FractionalCube.cubeLinedraw = function(a,b) {
	var N = Cube.distance(a,b);
	var results = [];
	var _g1 = 0;
	var _g = N + 1;
	while(_g1 < _g) {
		var i = _g1++;
		results.push(FractionalCube.cubeRound(FractionalCube.cubeLerp(a,b,1.0 / Math.max(1,N) * i)));
	}
	return results;
};
FractionalCube.prototype = {
	v: function() {
		return [this.x,this.y,this.z];
	}
	,toString: function() {
		return "#{" + this.v().join(",") + "}";
	}
};
var Grid = $hx_exports.Grid = function(scale,orientation,shape) {
	this.scale = scale;
	this.orientation = orientation;
	this.hexes = shape;
};
Grid.boundsOfPoints = function(points) {
	var minX = 0.0;
	var minY = 0.0;
	var maxX = 0.0;
	var maxY = 0.0;
	var _g = 0;
	while(_g < points.length) {
		var p = points[_g];
		++_g;
		if(p.x < minX) minX = p.x;
		if(p.x > maxX) maxX = p.x;
		if(p.y < minY) minY = p.y;
		if(p.y > maxY) maxY = p.y;
	}
	return { minX : minX, maxX : maxX, minY : minY, maxY : maxY};
};
Grid.twoAxisToCube = function(hex) {
	return new Cube(hex.q,-hex.r - hex.q,hex.r);
};
Grid.cubeToTwoAxis = function(cube) {
	return new Hex(cube.x | 0,cube.z | 0);
};
Grid.oddQToCube = function(hex) {
	var x = hex.q;
	var z = hex.r - (hex.q - (hex.q & 1) >> 1);
	return new Cube(x,-x - z,z);
};
Grid.cubeToOddQ = function(cube) {
	var x = cube.x | 0;
	var z = cube.z | 0;
	return new Hex(x,z + (x - (x & 1) >> 1));
};
Grid.evenQToCube = function(hex) {
	var x = hex.q;
	var z = hex.r - (hex.q + (hex.q & 1) >> 1);
	return new Cube(x,-x - z,z);
};
Grid.cubeToEvenQ = function(cube) {
	var x = cube.x | 0;
	var z = cube.z | 0;
	return new Hex(x,z + (x + (x & 1) >> 1));
};
Grid.oddRToCube = function(hex) {
	var z = hex.r;
	var x = hex.q - (hex.r - (hex.r & 1) >> 1);
	return new Cube(x,-x - z,z);
};
Grid.cubeToOddR = function(cube) {
	var x = cube.x | 0;
	var z = cube.z | 0;
	return new Hex(x + (z - (z & 1) >> 1),z);
};
Grid.evenRToCube = function(hex) {
	var z = hex.r;
	var x = hex.q - (hex.r + (hex.r & 1) >> 1);
	return new Cube(x,-x - z,z);
};
Grid.cubeToEvenR = function(cube) {
	var x = cube.x | 0;
	var z = cube.z | 0;
	return new Hex(x + (z + (z & 1) >> 1),z);
};
Grid.trapezoidalShape = function(minQ,maxQ,minR,maxR,toCube) {
	var hexes = [];
	var _g1 = minQ;
	var _g = maxQ + 1;
	while(_g1 < _g) {
		var q = _g1++;
		var _g3 = minR;
		var _g2 = maxR + 1;
		while(_g3 < _g2) {
			var r = _g3++;
			hexes.push(toCube(new Hex(q,r)));
		}
	}
	return hexes;
};
Grid.triangularShape = function(size) {
	var hexes = [];
	var _g1 = 0;
	var _g = size + 1;
	while(_g1 < _g) {
		var k = _g1++;
		var _g3 = 0;
		var _g2 = k + 1;
		while(_g3 < _g2) {
			var i = _g3++;
			hexes.push(new Cube(i,-k,k - i));
		}
	}
	return hexes;
};
Grid.hexagonalShape = function(size) {
	var hexes = [];
	var _g1 = -size;
	var _g = size + 1;
	while(_g1 < _g) {
		var x = _g1++;
		var _g3 = -size;
		var _g2 = size + 1;
		while(_g3 < _g2) {
			var y = _g3++;
			var z = -x - y;
			if(Math.abs(x) <= size && Math.abs(y) <= size && Math.abs(z) <= size) hexes.push(new Cube(x,y,z));
		}
	}
	return hexes;
};
Grid.prototype = {
	hexToCenter: function(cube) {
		var s;
		var size = this.scale / 2;
		if(this.orientation) s = new ScreenCoordinate(Math.sqrt(3) * cube.x + Math.sqrt(3) / 2 * cube.z,1.5 * cube.z); else s = new ScreenCoordinate(1.5 * cube.x,Math.sqrt(3) / 2 * cube.x + Math.sqrt(3) * cube.z);
		return s.scale(size);
	}
	,cartesianToHex: function(p) {
		var size = this.scale / 2;
		p = p.scale(1 / size);
		if(this.orientation) {
			var q = Math.sqrt(3) / 3 * p.x + -0.333333333333333315 * p.y;
			var r = 0.66666666666666663 * p.y;
			return new FractionalCube(q,-q - r,r);
		} else {
			var q1 = 0.66666666666666663 * p.x;
			var r1 = -0.333333333333333315 * p.x + Math.sqrt(3) / 3 * p.y;
			return new FractionalCube(q1,-q1 - r1,r1);
		}
	}
	,bounds: function() {
		var _g = this;
		var centers = Lambda.array(this.hexes.map(function(hex) {
			return _g.hexToCenter(hex);
		}));
		var b1 = Grid.boundsOfPoints(this.polygonVertices());
		var b2 = Grid.boundsOfPoints(centers);
		return { minX : b1.minX + b2.minX, maxX : b1.maxX + b2.maxX, minY : b1.minY + b2.minY, maxY : b1.maxY + b2.maxY};
	}
	,polygonVertices: function() {
		var points = [];
		var _g = 0;
		while(_g < 6) {
			var i = _g++;
			var angle;
			angle = 2 * Math.PI * (2 * i - (this.orientation?1:0)) / 12;
			points.push(new ScreenCoordinate(0.5 * this.scale * Math.cos(angle),0.5 * this.scale * Math.sin(angle)));
		}
		return points;
	}
};
var Hex = $hx_exports.Hex = function(q,r) {
	this.q = q;
	this.r = r;
};
Hex.prototype = {
	toString: function() {
		return this.q + ":" + this.r;
	}
};
var HxOverrides = function() { };
HxOverrides.iter = function(a) {
	return { cur : 0, arr : a, hasNext : function() {
		return this.cur < this.arr.length;
	}, next : function() {
		return this.arr[this.cur++];
	}};
};
var Lambda = function() { };
Lambda.array = function(it) {
	var a = [];
	var $it0 = $iterator(it)();
	while( $it0.hasNext() ) {
		var i = $it0.next();
		a.push(i);
	}
	return a;
};
var ScreenCoordinate = $hx_exports.ScreenCoordinate = function(x,y) {
	this.x = x;
	this.y = y;
};
ScreenCoordinate.prototype = {
	equals: function(p) {
		return this.x == p.x && this.y == p.y;
	}
	,toString: function() {
		return this.x + "," + this.y;
	}
	,length_squared: function() {
		return this.x * this.x + this.y * this.y;
	}
	,length: function() {
		return Math.sqrt(this.length_squared());
	}
	,normalize: function() {
		var d = this.length();
		return new ScreenCoordinate(this.x / d,this.y / d);
	}
	,scale: function(d) {
		return new ScreenCoordinate(this.x * d,this.y * d);
	}
	,rotateLeft: function() {
		return new ScreenCoordinate(this.y,-this.x);
	}
	,rotateRight: function() {
		return new ScreenCoordinate(-this.y,this.x);
	}
	,add: function(p) {
		return new ScreenCoordinate(this.x + p.x,this.y + p.y);
	}
	,subtract: function(p) {
		return new ScreenCoordinate(this.x - p.x,this.y - p.y);
	}
	,dot: function(p) {
		return this.x * p.x + this.y * p.y;
	}
	,cross: function(p) {
		return this.x * p.y - this.y * p.x;
	}
	,distance: function(p) {
		return this.subtract(p).length();
	}
};
var Std = function() { };
Std["int"] = function(x) {
	return x | 0;
};
function $iterator(o) { if( o instanceof Array ) return function() { return HxOverrides.iter(o); }; return typeof(o.iterator) == 'function' ? $bind(o,o.iterator) : o.iterator; }
var $_, $fid = 0;
function $bind(o,m) { if( m == null ) return null; if( m.__id__ == null ) m.__id__ = $fid++; var f; if( o.hx__closures__ == null ) o.hx__closures__ = {}; else f = o.hx__closures__[m.__id__]; if( f == null ) { f = function(){ return f.method.apply(f.scope, arguments); }; f.scope = o; f.method = m; o.hx__closures__[m.__id__] = f; } return f; }
if(Array.prototype.map == null) Array.prototype.map = function(f) {
	var a = [];
	var _g1 = 0;
	var _g = this.length;
	while(_g1 < _g) {
		var i = _g1++;
		a[i] = f(this[i]);
	}
	return a;
};
Cube.directions = [new Cube(1,-1,0),new Cube(1,0,-1),new Cube(0,1,-1),new Cube(-1,1,0),new Cube(-1,0,1),new Cube(0,-1,1)];
Cube.diagonals = [new Cube(2,-1,-1),new Cube(1,1,-2),new Cube(-1,2,-1),new Cube(-2,1,1),new Cube(-1,-1,2),new Cube(1,-2,1)];
Grid.SQRT_3_2 = Math.sqrt(3) / 2;
})(typeof console != "undefined" ? console : {log:function(){}}, typeof window != "undefined" ? window : exports);
