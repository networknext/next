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

import CustomScreenGridCellLayer from './CustomScreenGridCellLayer'
import CustomGridAggregationLayer from './CustomGridAggregationLayer'
import { getBoundingBox, getFloatTexture, getValueFunc, pointToDensityGridDataCPU } from './CustomGridAggregationUtils'

const defaultProps = {
  ...CustomScreenGridCellLayer.defaultProps,
  getPosition: { type: 'accessor', value: d => d.position },
  getWeight: { type: 'accessor', value: 1 },
  gpuAggregation: true,
  aggregation: 'SUM'
}

const POSITION_ATTRIBUTE_NAME = 'positions'
const DIMENSIONS = {
  data: {
    props: ['cellSizePixels']
  },
  weights: {
    props: ['aggregation'],
    accessors: ['getWeight']
  }
}

export default class CustomScreenGridLayer extends CustomGridAggregationLayer {
  constructor (props) {
    super(props)
    window.deck.log.once(`CustomScreenGridLayer: ${this.id} loaded`)()
  }

  initializeState () {
    const { gl } = this.context
    if (!CustomScreenGridCellLayer.isSupported(gl)) {
      // max aggregated value is sampled from a float texture
      this.setState({ supported: false })
      window.deck.log.error(`CustomScreenGridLayer: ${this.id} is not supported on this browser`)()
      return
    }
    super.initializeState({
      dimensions: DIMENSIONS,
      getCellSize: props => props.cellSizePixels
    })
    const weights = {
      count: {
        size: 1,
        operation: window.deck.AGGREGATION_OPERATION.SUM,
        needMax: true,
        maxTexture: getFloatTexture(gl, { id: `${this.id}-max-texture` })
      }
    }
    this.setState({
      supported: true,
      projectPoints: true, // aggregation in screen space
      weights,
      subLayerData: { attributes: {} },
      maxTexture: weights.count.maxTexture,
      positionAttributeName: 'positions',
      posOffset: [0, 0],
      translation: [1, -1]
    })
    const attributeManager = this.getAttributeManager()
    attributeManager.add({
      [POSITION_ATTRIBUTE_NAME]: {
        size: 3,
        accessor: 'getPosition',
        type: 0x140a,
        fp64: this.use64bitPositions()
      },
      // this attribute is used in gpu aggregation path only
      count: { size: 3, accessor: 'getWeight' }
    })
  }

  shouldUpdateState ({ changeFlags }) {
    return this.state.supported && changeFlags.somethingChanged
  }

  updateState (opts) {
    super.updateState(opts)
  }

  renderLayers () {
    if (!this.state.supported) {
      return []
    }
    const { maxTexture, numRow, numCol, weights } = this.state
    const { updateTriggers } = this.props
    const { aggregationBuffer } = weights.count
    const CellLayerClass = this.getSubLayerClass('cells', CustomScreenGridCellLayer)

    return new CellLayerClass(
      this.props,
      this.getSubLayerProps({
        id: 'cell-layer',
        updateTriggers
      }),
      {
        data: { attributes: { instanceCounts: aggregationBuffer } },
        maxTexture,
        numInstances: numRow * numCol
      }
    )
  }

  finalizeState () {
    super.finalizeState()

    const { aggregationBuffer, maxBuffer, maxTexture } = this.state

    if (aggregationBuffer) {
      aggregationBuffer.delete()
    }

    if (maxBuffer) {
      maxBuffer.delete()
    }

    if (maxTexture) {
      maxTexture.delete()
    }
  }

  getPickingInfo ({ info, mode }) {
    if (mode === 'hover') {
      return info
    }
    const { index } = info
    if (index >= 0) {
      const { gpuGridAggregator, gpuAggregation, weights } = this.state
      // Get count aggregation results
      const aggregationResults = gpuAggregation
        ? gpuGridAggregator.getData('count')
        : weights.count

      // Each instance (one cell) is aggregated into single pixel,
      // Get current instance's aggregation details.
      info.object = window.deck._GPUGridAggregator.getAggregationData({
        pixelIndex: index,
        ...aggregationResults
      })

      // perform CPU aggregation for full list of points for each cell
      // TODO: Move this to only happen on aggregations
      const { props } = this
      let { gridHash } = this.state
      if (!gridHash) {
        const { gridOffset, translation, projectPoints } = this.state
        const attributes = this.getAttributes()
        const boundingBox = getBoundingBox(attributes, this.getNumInstances())
        const { viewport } = this.context
        const cpuAggregation = pointToDensityGridDataCPU(props, {
          gridOffset,
          attributes,
          viewport,
          translation,
          boundingBox,
          projectPoints
        })
        gridHash = cpuAggregation.gridHash
      }
      const lookUpKey = this.getHashKeyForIndex(index)
      const lookUpObject = gridHash[lookUpKey]

      Object.assign(info.object, lookUpObject)
    }

    return info
  }

  getHashKeyForIndex (index) {
    const { numRow, numCol, gridOffset } = this.state
    const boundingBox = getBoundingBox(this.getAttributes(), this.getNumInstances())
    const gridSize = [numCol, numRow]
    const gridOrigin = [boundingBox.xMin, boundingBox.yMin]
    const cellSize = [gridOffset.xOffset, gridOffset.yOffset]

    const yIndex = Math.floor(index / gridSize[0])
    const xIndex = index - yIndex * gridSize[0]

    // This will match the index to the hash-key to access aggregation data from CPU aggregation results.
    const latIdx = Math.floor(
      (yIndex * cellSize[1] + 90 + cellSize[1] / 2) / cellSize[1]
    )
    const lonIdx = Math.floor(
      (xIndex * cellSize[0] + 180 + cellSize[0] / 2) / cellSize[0]
    )
    return `${latIdx}-${lonIdx}`
  }

  // Aggregation Overrides

  updateResults ({ aggregationData, maxData }) {
    const { count } = this.state.weights
    count.aggregationData = aggregationData
    count.aggregationBuffer.setData({ data: aggregationData })
    count.maxData = maxData
    count.maxTexture.setImageData({ data: maxData })
  }

  /* eslint-disable complexity, max-statements */
  updateAggregationState (opts) {
    const cellSize = opts.props.cellSizePixels
    const cellSizeChanged = opts.oldProps.cellSizePixels !== cellSize
    const { viewportChanged } = opts.changeFlags
    let gpuAggregation = opts.props.gpuAggregation
    if (this.state.gpuAggregation !== opts.props.gpuAggregation) {
      if (gpuAggregation && !window.deck._GPUGridAggregator.isSupported(this.context.gl)) {
        window.deck.log.warn('GPU Grid Aggregation not supported, falling back to CPU')()
        gpuAggregation = false
      }
    }
    const gpuAggregationChanged = gpuAggregation !== this.state.gpuAggregation
    this.setState({
      gpuAggregation
    })

    const positionsChanged = this.isAttributeChanged(POSITION_ATTRIBUTE_NAME)

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
  /* eslint-enable complexity, max-statements */

  // Private

  _updateAccessors (opts) {
    const { getWeight, aggregation, data } = opts.props
    const { count } = this.state.weights
    if (count) {
      count.getWeight = getWeight
      count.operation = window.deck.AGGREGATION_OPERATION[aggregation]
    }
    this.setState({ getValue: getValueFunc(aggregation, getWeight, { data }) })
  }

  _resetResults () {
    const { count } = this.state.weights
    if (count) {
      count.aggregationData = null
    }
  }
}

CustomScreenGridLayer.layerName = 'CustomScreenGridLayer'
CustomScreenGridLayer.defaultProps = defaultProps
