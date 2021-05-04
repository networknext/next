import { ScreenGridLayer, _GPUGridAggregator } from '@deck.gl/aggregation-layers'
import { log, createIterable } from '@deck.gl/core'

export default class CustomScreenGridLayer extends ScreenGridLayer {
  constructor (props: any) {
    super(props)
    log.log(1, 'CustomScreenGridLayer: loaded')()
  }

  updateAggregationState (opts: any) {
    const cellSize = opts.props.cellSizePixels
    const cellSizeChanged = opts.oldProps.cellSizePixels !== cellSize
    const { viewportChanged } = opts.changeFlags
    let gpuAggregation = opts.props.gpuAggregation
    if (this.state.gpuAggregation !== opts.props.gpuAggregation) {
      if (gpuAggregation && !_GPUGridAggregator.isSupported(this.context.gl)) {
        log.warn('GPU Grid Aggregation not supported, falling back to CPU')()
        gpuAggregation = false
      }
    }
    const gpuAggregationChanged = gpuAggregation !== this.state.gpuAggregation
    this.setState({
      gpuAggregation
    })

    const positionsChanged = this.isAttributeChanged('positions')

    let { boundingBox } = this.state
    if (positionsChanged) {
      boundingBox = this.getBoundingBox(this.getAttributes(), this.getNumInstances())
      this.setState({ boundingBox })
    }

    const { dimensions } = this.state
    const { data, weights } = dimensions
    const aggregationDataDirty =
      positionsChanged ||
      gpuAggregationChanged ||
      viewportChanged ||
      this.isAggregationDirty(opts, {
        compareAll: gpuAggregation, // check for all (including extentions props) when using gpu aggregation
        dimension: data
      })
    const aggregationWeightsDirty = this.isAggregationDirty(opts, { dimension: weights })

    this.setState({
      aggregationDataDirty,
      aggregationWeightsDirty
    })

    const { viewport } = this.context

    if (viewportChanged || cellSizeChanged) {
      const { width, height } = viewport
      const numCol = Math.ceil(width / cellSize)
      const numRow = Math.ceil(height / cellSize)
      this.allocateResources(numRow, numCol)
      this.setState({
        // transformation from clipspace to screen(pixel) space
        scaling: [width / 2, -height / 2, 1],
        gridOffset: { xOffset: cellSize, yOffset: cellSize },
        width,
        height,
        numCol,
        numRow
      })
    }

    if (aggregationWeightsDirty) {
      this._updateAccessors(opts)
    }
    if (aggregationDataDirty || aggregationWeightsDirty) {
      this._resetResults()
    }
  }

  updateState (opts: any) {
    super.updateState(opts)
    const { aggregationDirty } = this.state
    if (aggregationDirty) {
      // reset cached CPU Aggregation results (used for picking)
      this.setState({
        gridHash: null
      })
    }
  }

  getPickingInfo ({ info, mode }: any) {
    if (mode === 'query') {
      console.log(this)
    }

    const { index } = info
    const object = null

    if (index >= 0 && mode !== 'hover') {
      // perform CPU aggregation for full list of points for each cell
      const { props } = this
      let { gridHash } = this.state
      if (!gridHash) {
        const { gridOffset, translation, boundingBox } = this.state
        const { viewport } = this.context
        const attributes = this.getAttributes()
        console.log(attributes)
        const cpuAggregation = this.pointToDensityGridDataCPU(props, {
          gridOffset,
          attributes,
          viewport,
          translation,
          boundingBox
        })
        console.log('cpuAggregation')
        console.log(cpuAggregation)
        gridHash = cpuAggregation.gridHash || {}
        this.setState({ gridHash })
      }
      const key = this.getHashKeyForIndex(index)
      console.log('key')
      console.log(key)
      const cpuAggregationData = this.state.gridHash[key]
      console.log(cpuAggregationData)
      Object.assign(object, cpuAggregationData)
    }

    info.picked = Boolean(object)
    info.object = object

    return info
  }

  getHashKeyForIndex (index: any) {
    const { numRow, numCol, boundingBox, gridOffset } = this.state
    const gridSize = [numCol, numRow]
    const gridOrigin = [boundingBox.xMin, boundingBox.yMin]
    const cellSize = [gridOffset.xOffset, gridOffset.yOffset]

    const yIndex = Math.floor(index / gridSize[0])
    const xIndex = index - yIndex * gridSize[0]
    // This will match the index to the hash-key to access aggregation data from CPU aggregation results.
    const latIdx = Math.floor(
      (yIndex * cellSize[1] + gridOrigin[1] + 90 + cellSize[1] / 2) / cellSize[1]
    )
    const lonIdx = Math.floor(
      (xIndex * cellSize[0] + gridOrigin[0] + 180 + cellSize[0] / 2) / cellSize[0]
    )
    return `${latIdx}-${lonIdx}`
  }

  getPositionForIndex (index: any) {
    const { numRow, numCol, boundingBox, gridOffset } = this.state
    const gridSize = [numCol, numRow]
    const gridOrigin = [boundingBox.xMin, boundingBox.yMin]
    const cellSize = [gridOffset.xOffset, gridOffset.yOffset]

    const yIndex = Math.floor(index / gridSize[0])
    const xIndex = index - yIndex * gridSize[0]
    const yPos = yIndex * cellSize[1] + gridOrigin[1]
    const xPos = xIndex * cellSize[0] + gridOrigin[0]
    return [xPos, yPos]
  }

  getBoundingBox (attributes: any, vertexCount: any) {
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
      xMin: this.toFinite(xMin),
      xMax: this.toFinite(xMax),
      yMin: this.toFinite(yMin),
      yMax: this.toFinite(yMax)
    }

    return boundingBox
  }

  toFinite (n: number) {
    return Number.isFinite(n) ? n : 0
  }

  pointToDensityGridDataCPU (props: any, aggregationParams: any) {
    const hashInfo = this.pointsToGridHashing(props, aggregationParams)
    const result = this.getGridLayerDataFromGridHash(hashInfo)

    console.log('hashInfo')
    console.log(hashInfo)
    console.log('result')
    console.log(result)

    return {
      gridHash: hashInfo.gridHash,
      gridOffset: hashInfo.gridOffset,
      data: result
    }
  }

  pointsToGridHashing (props: any, aggregationParams: any) {
    // TODO: cellsize is not of the right type. We want cell size in KM not pixels
    console.log('props')
    console.log(props)
    console.log('props')
    const { data = [], cellSize } = props
    const { attributes, viewport, projectPoints, numInstances } = aggregationParams
    const positions = attributes.positions.value
    const { size } = attributes.positions.getAccessor()
    const boundingBox =
      aggregationParams.boundingBox || this.getPositionBoundingBox(attributes.positions, numInstances)
    const offsets = aggregationParams.posOffset || [180, 90]
    const gridOffset = aggregationParams.gridOffset || this.getGridOffset(boundingBox, cellSize)

    if (gridOffset.xOffset <= 0 || gridOffset.yOffset <= 0) {
      return { gridHash: {}, gridOffset }
    }

    const { width, height } = viewport
    const numCol = Math.ceil(width / gridOffset.xOffset)
    const numRow = Math.ceil(height / gridOffset.yOffset)

    // calculate count per cell
    const gridHash: any = {}

    const { iterable, objectInfo } = createIterable(data)
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

  getGridLayerDataFromGridHash ({ gridHash, gridOffset, offsets }: any) {
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

  // Calculate bounding box of position attribute
  getPositionBoundingBox (positionAttribute: any, numInstance: any) {
    // TODO - value might not exist (e.g. attribute transition)
    const positions = positionAttribute.value
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
}
