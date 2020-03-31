window.JSONRPCClient = {

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