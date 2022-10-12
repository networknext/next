// Copyright (c) 2015 - 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

const DEFAULT_PARAMETERS = {
  0x2800: 0x2600,
  0x2801: 0x2600
}

export const defaultColorRange = [
  [255, 255, 178],
  [254, 217, 118],
  [254, 178, 76],
  [253, 141, 60],
  [240, 59, 32],
  [189, 0, 38]
]

const R_EARTH = 6378000

function toFinite (n) {
  return Number.isFinite(n) ? n : 0
}

export function getFloatTexture (gl, opts = {}) {
  const {
    width = 1,
    height = 1,
    data = null,
    unpackFlipY = true,
    parameters = DEFAULT_PARAMETERS
  } = opts
  const texture = new window.luma.Texture2D(gl, {
    data,
    format: window.luma.isWebGL2(gl) ? 0x8814 : 0x1908,
    type: 0x1406,
    border: 0,
    mipmaps: false,
    parameters,
    dataFormat: 0x1908,
    width,
    height,
    unpackFlipY
  })
  return texture
}

// Parse input data to build positions, wights and bounding box.
/* eslint-disable max-statements */
export function getBoundingBox (attributes, vertexCount) {
  // TODO - value might not exist (e.g. attribute transition)
  const positions = attributes.positions.value

  let yMin = Infinity
  let yMax = -Infinity
  let xMin = Infinity
  let xMax = -Infinity
  let y
  let x

  for (let i = 0; i < vertexCount; i++) {
    x = positions[i * 3]
    y = positions[i * 3 + 1]
    yMin = y < yMin ? y : yMin
    yMax = y > yMax ? y : yMax
    xMin = x < xMin ? x : xMin
    xMax = x > xMax ? x : xMax
  }

  const boundingBox = {
    xMin: toFinite(xMin),
    xMax: toFinite(xMax),
    yMin: toFinite(yMin),
    yMax: toFinite(yMax)
  }

  return boundingBox
}

// Aligns `inValue` to given `cellSize`
export function alignToCell (inValue, cellSize) {
  const sign = inValue < 0 ? -1 : 1

  let value = sign < 0 ? Math.abs(inValue) + cellSize : Math.abs(inValue)

  value = Math.floor(value / cellSize) * cellSize

  return value * sign
}

/* eslint-enable max-statements */

// Returns XY translation for positions to peform aggregation in +ve sapce
function getTranslation (boundingBox, gridOffset, coordinateSystem, viewport) {
  const { width, height } = viewport

  // Origin to define grid
  // DEFAULT coordinate system is treated as LNGLAT
  const worldOrigin =
    coordinateSystem === window.deck.COORDINATE_SYSTEM.CARTESIAN ? [-width / 2, -height / 2] : [-180, -90]

  // Other coordinate systems not supported/verified yet
  window.deck.log.assert(
    coordinateSystem === window.deck.COORDINATE_SYSTEM.CARTESIAN ||
      coordinateSystem === window.deck.COORDINATE_SYSTEM.LNGLAT ||
      coordinateSystem === window.deck.COORDINATE_SYSTEM.DEFAULT
  )

  const { xMin, yMin } = boundingBox
  return [
    // Align origin to match grid cell boundaries in CPU and GPU aggregations
    -1 * (alignToCell(xMin - worldOrigin[0], gridOffset.xOffset) + worldOrigin[0]),
    -1 * (alignToCell(yMin - worldOrigin[1], gridOffset.yOffset) + worldOrigin[1])
  ]
}

/**
 * with a given x-km change, calculate the increment of latitude
 * based on stackoverflow http://stackoverflow.com/questions/7477003
 * @param {number} dy - change in km
 * @return {number} - increment in latitude
 */
function calculateLatOffset (dy) {
  return (dy / R_EARTH) * (180 / Math.PI)
}

/**
 * with a given x-km change, and current latitude
 * calculate the increment of longitude
 * based on stackoverflow http://stackoverflow.com/questions/7477003
 * @param {number} lat - latitude of current location (based on city)
 * @param {number} dx - change in km
 * @return {number} - increment in longitude
 */
function calculateLonOffset (lat, dx) {
  return ((dx / R_EARTH) * (180 / Math.PI)) / Math.cos((lat * Math.PI) / 180)
}

/**
 * calculate grid layer cell size in lat lon based on world unit size
 * and current latitude
 * @param {number} cellSize
 * @param {number} latitude
 * @returns {object} - lat delta and lon delta
 */
function calculateGridLatLonOffset (cellSize, latitude) {
  const yOffset = calculateLatOffset(cellSize)
  const xOffset = calculateLonOffset(latitude, cellSize)
  return { yOffset, xOffset }
}

/**
 * Based on geometric center of sample points, calculate cellSize in lng/lat (degree) space
 * @param {object} boundingBox - {xMin, yMin, xMax, yMax} contains bounding box of data
 * @param {number} cellSize - grid cell size in meters
 * @param {boolean, optional} converToDegrees - when true offsets are converted from meters to lng/lat (degree) space
 * @returns {xOffset, yOffset} - cellSize size
 */

export function getGridOffset (boundingBox, cellSize, convertToMeters = true) {
  if (!convertToMeters) {
    return { xOffset: cellSize, yOffset: cellSize }
  }

  const { yMin, yMax } = boundingBox
  const centerLat = (yMin + yMax) / 2

  return calculateGridLatLonOffset(cellSize, centerLat)
}

export function getGridParams (boundingBox, cellSize, viewport, coordinateSystem) {
  const gridOffset = getGridOffset(
    boundingBox,
    cellSize,
    coordinateSystem !== window.deck.COORDINATE_SYSTEM.CARTESIAN
  )

  const translation = getTranslation(boundingBox, gridOffset, coordinateSystem, viewport)

  const { xMin, yMin, xMax, yMax } = boundingBox

  const width = xMax - xMin + gridOffset.xOffset
  const height = yMax - yMin + gridOffset.yOffset

  const numCol = Math.ceil(width / gridOffset.xOffset)
  const numRow = Math.ceil(height / gridOffset.yOffset)
  return { gridOffset, translation, width, height, numCol, numRow }
}

// Calculate bounding box of position attribute
function getPositionBoundingBox (positionAttribute, numInstance) {
  // TODO - value might not exist (e.g. attribute transition)
  const positions = positionAttribute.source.value
  const { size } = positionAttribute.getAccessor()

  let yMin = Infinity
  let yMax = -Infinity
  let xMin = Infinity
  let xMax = -Infinity
  let y
  let x

  for (let i = 0; i < numInstance; i++) {
    x = positions[i * size]
    y = positions[i * size + 1]
    if (Number.isFinite(x) && Number.isFinite(y)) {
      yMin = y < yMin ? y : yMin
      yMax = y > yMax ? y : yMax
      xMin = x < xMin ? x : xMin
      xMax = x > xMax ? x : xMax
    }
  }

  return { xMin, xMax, yMin, yMax }
}

export function PixelsToMeters (viewport) {
  return viewport.distanceScales.metersPerUnit[0]
}

/**
 * Project points into each cell, return a hash table of cells
 * @param {Iterable} points
 * @param {number} cellSize - unit size in meters
 * @param {function} getPosition - position accessor
 * @returns {object} - grid hash and cell dimension
 */
/* eslint-disable max-statements, complexity */
function pointsToGridHashing (props, aggregationParams) {
  const { data = [], cellSize } = props
  const { attributes, viewport, projectPoints, numInstances } = aggregationParams
  const positions = attributes.positions.value
  const { size } = attributes.positions.getAccessor()
  const boundingBox =
    aggregationParams.boundingBox || getPositionBoundingBox(attributes.positions, numInstances)
  const offsets = aggregationParams.posOffset || [180, 90]
  const gridOffset = aggregationParams.gridOffset || getGridOffset(boundingBox, cellSize)

  if (gridOffset.xOffset <= 0 || gridOffset.yOffset <= 0) {
    return { gridHash: {}, gridOffset }
  }

  const { width, height } = viewport
  const numCol = Math.ceil(width / gridOffset.xOffset)
  const numRow = Math.ceil(height / gridOffset.yOffset)

  // calculate count per cell
  const gridHash = {}

  const { iterable, objectInfo } = window.deck.createIterable(data)
  const position = new Array(3)
  for (const pt of iterable) {
    objectInfo.index++
    position[0] = positions[objectInfo.index * size]
    position[1] = positions[objectInfo.index * size + 1]
    position[2] = size >= 3 ? positions[objectInfo.index * size + 2] : 0

    const [x, y] = projectPoints ? viewport.project(position) : position
    if (Number.isFinite(x) && Number.isFinite(y)) {
      const yIndex = Math.floor((y + offsets[1]) / gridOffset.yOffset)
      const xIndex = Math.floor((x + offsets[0]) / gridOffset.xOffset)
      if (
        !projectPoints ||
        // when doing screen space agggregation (projectPoints = true), filter points outside of the viewport range.
        (xIndex >= 0 && xIndex < numCol && yIndex >= 0 && yIndex < numRow)
      ) {
        const key = `${yIndex}-${xIndex}`

        gridHash[key] = gridHash[key] || { count: 0, points: [], lonIdx: xIndex, latIdx: yIndex }
        gridHash[key].count += 1
        gridHash[key].points.push({
          source: pt,
          index: objectInfo.index
        })
      }
    }
  }

  return { gridHash, gridOffset, offsets: [offsets[0] * -1, offsets[1] * -1] }
}
/* eslint-enable max-statements, complexity */

function getGridLayerDataFromGridHash ({ gridHash, gridOffset, offsets }) {
  const data = new Array(Object.keys(gridHash).length)
  let i = 0
  for (const key in gridHash) {
    const idxs = key.split('-')
    const latIdx = parseInt(idxs[0], 10)
    const lonIdx = parseInt(idxs[1], 10)
    const index = i++

    data[index] = {
      index,
      position: [
        offsets[0] + gridOffset.xOffset * lonIdx,
        offsets[1] + gridOffset.yOffset * latIdx
      ],
      ...gridHash[key]
    }
  }
  return data
}

/**
 * Calculate density grid from an array of points
 * @param {Object} props - object containing :
 * @param {Iterable} [props.data] - data objects to be aggregated
 * @param {Integer} [props.cellSize] - size of the grid cell
 *
 * @param {Object} aggregationParams - object containing :
 * @param {Object} gridOffset - {xOffset, yOffset} cell size in meters
 * @param {Integer} width - width of the grid
 * @param {Integer} height - height of the grid
 * @param {Boolean} projectPoints - `true` if doing screen space projection, `false` otherwise
 * @param {Array} attributes - attributes array containing position values
 * @param {Viewport} viewport - viewport to be used for projection
 * @param {Array} posOffset - [xOffset, yOffset] offset to be applied to positions to get cell index
 * @param {Object} boundingBox - {xMin, yMin, xMax, yMax} bounding box of input data
 *
 * @returns {object} - grid data, cell dimension
 */
export function pointToDensityGridDataCPU (props, aggregationParams) {
  const hashInfo = pointsToGridHashing(props, aggregationParams)
  const result = getGridLayerDataFromGridHash(hashInfo)

  return {
    gridHash: hashInfo.gridHash,
    gridOffset: hashInfo.gridOffset,
    data: result
  }
}

function sumReducer (accu, cur) {
  return accu + cur
}

function maxReducer (accu, cur) {
  return cur > accu ? cur : accu
}

function minReducer (accu, cur) {
  return cur < accu ? cur : accu
}

export function getMean (pts, accessor) {
  if (Number.isFinite(accessor)) {
    return pts.length ? accessor : null
  }
  const filtered = pts.map(accessor).filter(Number.isFinite)

  return filtered.length ? filtered.reduce(sumReducer, 0) / filtered.length : null
}

export function getSum (pts, accessor) {
  if (Number.isFinite(accessor)) {
    return pts.length ? pts.length * accessor : null
  }
  const filtered = pts.map(accessor).filter(Number.isFinite)

  return filtered.length ? filtered.reduce(sumReducer, 0) : null
}

export function getMax (pts, accessor) {
  if (Number.isFinite(accessor)) {
    return pts.length ? accessor : null
  }
  const filtered = pts.map(accessor).filter(Number.isFinite)

  return filtered.length ? filtered.reduce(maxReducer, -Infinity) : null
}

export function getMin (pts, accessor) {
  if (Number.isFinite(accessor)) {
    return pts.length ? accessor : null
  }
  const filtered = pts.map(accessor).filter(Number.isFinite)

  return filtered.length ? filtered.reduce(minReducer, Infinity) : null
}

function wrapAccessor (accessor, context = {}) {
  if (Number.isFinite(accessor)) {
    return accessor
  }
  return pt => {
    context.index = pt.index
    return accessor(pt.source, context)
  }
}

export function wrapGetValueFunc (getValue, context = {}) {
  return pts => {
    context.indices = pts.map(pt => pt.index)
    return getValue(pts.map(pt => pt.source), context)
  }
}

// Function to convert from aggregation/accessor props (like colorAggregation and getColorWeight) to getValue prop (like getColorValue)
export function getValueFunc (aggregation, accessor, context) {
  const op = window.deck.AGGREGATION_OPERATION[aggregation] || window.deck.AGGREGATION_OPERATION.SUM
  accessor = wrapAccessor(accessor, context)
  switch (op) {
    case window.deck.AGGREGATION_OPERATION.MIN:
      return pts => getMin(pts, accessor)
    case window.deck.AGGREGATION_OPERATION.SUM:
      return pts => getSum(pts, accessor)
    case window.deck.AGGREGATION_OPERATION.MEAN:
      return pts => getMean(pts, accessor)
    case window.deck.AGGREGATION_OPERATION.MAX:
      return pts => getMax(pts, accessor)
    default:
      return null
  }
}

// Converts a colorRange array to a flat array with 4 components per color
export function colorRangeToFlatArray (colorRange, normalize = false, ArrayType = Float32Array) {
  let flatArray

  if (Number.isFinite(colorRange[0])) {
    // its already a flat array.
    flatArray = new ArrayType(colorRange)
  } else {
    // flatten it
    flatArray = new ArrayType(colorRange.length * 4)
    let index = 0

    for (let i = 0; i < colorRange.length; i++) {
      const color = colorRange[i]
      flatArray[index++] = color[0]
      flatArray[index++] = color[1]
      flatArray[index++] = color[2]
      flatArray[index++] = Number.isFinite(color[3]) ? color[3] : 255
    }
  }

  if (normalize) {
    for (let i = 0; i < flatArray.length; i++) {
      flatArray[i] /= 255
    }
  }
  return flatArray
}
