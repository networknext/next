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

import { pointToDensityGridDataCPU } from './CustomGridAggregationUtils'

export default class CustomGridAggregationLayer extends window.deck._AggregationLayer {
  initializeState ({ dimensions }) {
    const { gl } = this.context
    super.initializeState(dimensions)
    this.setState({
      // CPU aggregation results
      layerData: {},
      gpuGridAggregator: new window.deck._GPUGridAggregator(gl, { id: `${this.id}-gpu-aggregator` }),
      cpuGridAggregator: pointToDensityGridDataCPU
    })
  }

  updateState (opts) {
    // get current attributes
    super.updateState(opts)

    this.updateAggregationState(opts)

    const { aggregationDataDirty, aggregationWeightsDirty, gpuAggregation } = this.state
    if (this.getNumInstances() <= 0) {
      return
    }
    let aggregationDirty = false
    // CPU aggregation is two steps
    // 1. Create bins (based on cellSize and position) 2. Aggregate weights for each bin
    // For GPU aggregation both above steps are combined into one step

    // step-1
    if (aggregationDataDirty || (gpuAggregation && aggregationWeightsDirty)) {
      this._updateAggregation(opts)
      aggregationDirty = true
    }
    // step-2 (Applicalbe for CPU aggregation only)
    if (!gpuAggregation && (aggregationDataDirty || aggregationWeightsDirty)) {
      this._updateWeightBins()
      this._uploadAggregationResults()
      aggregationDirty = true
    }

    this.setState({ aggregationDirty })
  }

  finalizeState () {
    const { count } = this.state.weights
    if (count && count.aggregationBuffer) {
      count.aggregationBuffer.delete()
    }
    if (this.state.gpuGridAggregator) {
      this.state.gpuGridAggregator.delete()
    }
    super.finalizeState()
  }

  updateShaders (shaders) {
    if (this.state.gpuAggregation) {
      this.state.gpuGridAggregator.updateShaders(shaders)
    }
  }

  // Methods that can be overriden by subclasses for customizations

  updateAggregationState () {
    // Sublayers should implement this method.
    window.deck.log.assert(false)
  }

  allocateResources (numRow, numCol) {
    if (this.state.numRow !== numRow || this.state.numCol !== numCol) {
      const dataBytes = numCol * numRow * 4 * 4
      const gl = this.context.gl
      const { weights } = this.state
      for (const name in weights) {
        const weight = weights[name]
        if (weight.aggregationBuffer) {
          weight.aggregationBuffer.delete()
        }
        weight.aggregationBuffer = new window.luma.Buffer(gl, {
          byteLength: dataBytes,
          accessor: {
            size: 4,
            type: 0x1406,
            divisor: 1
          }
        })
      }
    }
  }

  updateResults ({ aggregationData, maxMinData, maxData, minData }) {
    const { count } = this.state.weights
    if (count) {
      count.aggregationData = aggregationData
      count.maxMinData = maxMinData
      count.maxData = maxData
      count.minData = minData
    }
  }

  // Private

  _updateAggregation (opts) {
    const {
      cpuGridAggregator,
      gpuGridAggregator,
      gridOffset,
      posOffset,
      translation = [0, 0],
      scaling = [0, 0, 0],
      boundingBox,
      projectPoints,
      gpuAggregation,
      numCol,
      numRow
    } = this.state
    const { props } = opts
    const { viewport } = this.context
    const attributes = this.getAttributes()
    const vertexCount = this.getNumInstances()

    if (!gpuAggregation) {
      const result = cpuGridAggregator(props, {
        gridOffset,
        projectPoints,
        attributes,
        viewport,
        posOffset,
        boundingBox
      })
      this.setState({
        layerData: result
      })
    } else {
      const { weights } = this.state
      gpuGridAggregator.run({
        weights,
        cellSize: [gridOffset.xOffset, gridOffset.yOffset],
        numCol,
        numRow,
        translation,
        scaling,
        vertexCount,
        projectPoints,
        attributes,
        moduleSettings: this.getModuleSettings()
      })
    }
  }

  _updateWeightBins () {
    const { getValue } = this.state

    const sortedBins = new window.deck._BinSorter(this.state.layerData.data || [], { getValue })
    this.setState({ sortedBins })
  }

  _uploadAggregationResults () {
    const { numCol, numRow } = this.state
    const { data } = this.state.layerData
    const { aggregatedBins, minValue, maxValue, totalCount } = this.state.sortedBins

    const ELEMENTCOUNT = 4
    const aggregationSize = numCol * numRow * ELEMENTCOUNT
    const aggregationData = new Float32Array(aggregationSize).fill(0)
    for (const bin of aggregatedBins) {
      const { lonIdx, latIdx } = data[bin.i]
      const { value, counts } = bin
      const cellIndex = (lonIdx + latIdx * numCol) * ELEMENTCOUNT
      aggregationData[cellIndex] = value
      aggregationData[cellIndex + ELEMENTCOUNT - 1] = counts
    }
    const maxMinData = new Float32Array([maxValue, 0, 0, minValue])
    const maxData = new Float32Array([maxValue, 0, 0, totalCount])
    const minData = new Float32Array([minValue, 0, 0, totalCount])
    this.updateResults({ aggregationData, maxMinData, maxData, minData })
  }
}

CustomGridAggregationLayer.layerName = 'CustomGridAggregationLayer'
