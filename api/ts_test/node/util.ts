import { strict as assert } from 'node:assert';
import { Client } from './client.js';

export class TestClient {
	client: Client;

	constructor() {
		process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = '0';

		assert(process.env['BASE_URL']);

		this.client = new Client(process.env['BASE_URL']);
	}
}
