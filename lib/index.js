const { getSdk } = require('balena-sdk');
const debug = require('debug')('name');
const fs = require("fs").promises;
const sdk = getSdk();
const BALENA_API_KEY = process.env.BALENA_API_KEY;
const BALENA_DEVICE_UUID = process.env.BALENA_DEVICE_UUID;
const CATEGORY = process.env.CATEGORY || 'season';

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
    const version = await fs.readFile('VERSION');
    return version;
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

const generateName = async function (category) {
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
            default:
                break;
        }
        return category;
    })
    .then((category) => {
        debug(`category set to ${category}.`);
        generateName(category)
            .then((name) => {
                debug(`generated name: ${name}`);
                rename(name);
            })
    })
    .catch((err) => console.error('failed ', err));
