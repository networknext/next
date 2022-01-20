jest.setTimeout(15000)

class DeckMock {
  public setProps () {
    console.log('Setting mock props')
  }
}

class MapMock {}

class LayerMock {}

class _AggregationLayerMock {}

// Take care of tests complaining about CDN imported map libs *shudders* -> this can go away when Safari starts support GL2...
(window as any).deck = {
  Deck: DeckMock,
  Layer: LayerMock,
  _AggregationLayer: _AggregationLayerMock,
  log: {
    once: () => {
      return console.log
    }
  }
};

(window as any).mapboxgl = {
  Map: MapMock
}
