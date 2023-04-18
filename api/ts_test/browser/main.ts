import * as client from './client.js';

import * as testGet from './test-get.js';

const mods = [
	testGet,
];

document.body.style.fontFamily = 'monospace';

const c = new client.Client('..');

const proms: Map<string, Promise<void>> = new Map();

for (const mod of mods) {
	for (const [funcName, func] of Object.entries(mod)) {
		if (!funcName.startsWith('test') || typeof func != 'function') {
			continue;
		}

		proms.set(funcName, func(c));
	}
}

for (const [funcName, prom] of proms) {
	const div = document.createElement('div');
	document.body.appendChild(div);

	div.innerText = `${funcName}...`;

	const result = document.createElement('span');
	div.appendChild(result);
	result.style.whiteSpace = 'pre';

	try {
		await prom;

		result.innerText = 'PASS';
		result.style.color = 'green';
	} catch (e) {
		result.innerText = `${e}`;
		result.style.color = 'red';
	}
}
