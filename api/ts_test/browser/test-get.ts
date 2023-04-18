import * as client from './client.js';

export async function testGet(c: client.Client) {
	console.log(await c.debugInfo());

	throw 'foo';
}
