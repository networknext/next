jest.setTimeout(15000);

// Take care of tests complaining about CDN imported map libs *shudders* -> this can go away when Safari starts support GL2...
(window as any).deck = {
  Layer: null,
  _AggregationLayer: null
}
