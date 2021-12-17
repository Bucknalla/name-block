const fs = require("fs").promises;
const debug = require('debug')('sort');
const glob = require("glob");

const getDirectories = function (src, callback) {
    glob(src + '/**/*.txt', callback);
};

const sortList = function (list) {
    const words = list.split('\n');
    const unique = Array.from(new Set(words)).map(name => name.toLowerCase());
    const sorted = unique.sort()
    return sorted;
}

getDirectories('data', function (err, res) {
    res.map(async file => {
        const sort = await fs.readFile(file, "utf-8")
            .then(
                (data) => {
                    return sortList(data);
                }
            )
            .then(
                async (sorted) => {
                    await fs.writeFile(`${file}`, sorted.join("\n"));
                })
            .catch((err) => console.error('failed to read file', err));
        return sort;
    })
})
