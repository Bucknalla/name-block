const { getSdk } = require('balena-sdk');
const debug = require('debug')('name');
const fs = require("fs").promises;
var pjson = require('./package.json');
var https = require('https');
const sdk = getSdk();
const BALENA_API_KEY = process.env.BALENA_API_KEY;
const BALENA_DEVICE_UUID = process.env.BALENA_DEVICE_UUID;
const CATEGORY = process.env.CATEGORY || 'season';

const getURLs = async function () {
    return process.env.URL.split(',');
}

const getMonth = function () {
    const date = new Date();
    const month = date.getMonth();
    switch (true) {
        case (10 <= month || month < 2):
            return 'winter'
        case (2 <= month && month < 5):
            return 'spring'
        case (5 <= month && month < 8):
            return 'summer'
        case (8 <= month && month < 10):
            return 'autumn'
    }
}

const getVerison = async function () {
    return pjson.version;
}

const rename = async function (name) {
    try {
        await sdk.auth.logout();
        await sdk.auth.loginWithToken(BALENA_API_KEY);
        await sdk.models.device.envVar.get(BALENA_DEVICE_UUID, 'NAME')
            .then(function (value) {
                if (typeof value === 'undefined') {
                    debug(`setting device envar to ${name}`)
                    sdk.models.device.envVar.set(BALENA_DEVICE_UUID, 'NAME', name)
                        .then(() => {
                            debug(`generated name: ${name}`);
                            return sdk.models.device.rename(BALENA_DEVICE_UUID, name);
                        })
                } else {
                    debug(`failed, device name already set.`)
                }
            });
    } catch (error) {
        debug(error);
    }
}

const randomWord = function (words) {
    const lines = words.split('\n');
    const word = lines[Math.floor(Math.random() * lines.length)];
    return word;
}

function httpRequest(params,) {
    return new Promise(function (resolve, reject) {
        var req = https.request(params, function (res) {
            if (res.statusCode < 200 || res.statusCode >= 300) {
                return reject(new Error('statusCode=' + res.statusCode));
            }
            var body = [];
            res.on('data', function (chunk) {
                body.push(chunk);
            });
            res.on('end', function () {
                try {
                    body = Buffer.concat(body).toString();
                } catch (e) {
                    reject(e);
                }
                resolve(body);
            });
        });
        req.on('error', function (err) {
            reject(err);
        });
        req.end();
    });
}

const generateName = async function (category) {
    if (category !== 'url') {
        const adj = await fs.readFile(`data/${category}/adj.txt`, "utf-8")
            .then(
                (data) => {
                    return randomWord(data);
                }
            )
            .catch((err) => console.error('failed to read file', err));

        const noun = await fs.readFile(`data/${category}/noun.txt`, "utf-8")
            .then(
                (data) => {
                    return randomWord(data);
                }
            )
            .catch((err) => console.error('failed to read file', err));
        return (`${adj}-${noun}`);
    }
    else {
        const urls = await getURLs()
            .catch((err) => console.error("Envar 'url' not set", err));
        const noun = await httpRequest(urls[0])
            .then((data) => {
                return randomWord(data);
            });

        if (typeof urls[1] !== 'undefined') {
            const adj = await httpRequest(urls[1])
                .then((data) => {
                    return randomWord(data);
                });
            return (`${adj}-${noun}`);
        }
        else {
            return (`${noun}`);
        }
    }

}

let category = CATEGORY;
getVerison()
    .then((version) => {
        debug(`name-block: v${version}`)
    })
    .then(() => {
        if (typeof category === 'undefined') {
            throw new Error();
        }
        switch (category) {
            case 'undefined':
                throw new Error();
            case 'season':
                category = getMonth();
                break;
            case 'url':
                break;
            case 'rude':
                break;
            case 'animals':
                break;
            default:
                category = 'animals';
                break;
        }
        return category;
    })
    .then((category) => {
        debug(`category set to ${category}.`);
        generateName(category)
            .then((name) => {
                rename(name.toLowerCase());
            })
    })
    .catch((err) => console.error('failed ', err));
