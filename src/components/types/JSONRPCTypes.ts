export interface JSONRPCError {
  message: string;
  code: number;
  data: any;
}

export interface JSONRPCJSON {
  jsonrpc: string;
  id: string;
  response?: any;
  error?: JSONRPCError;
}
