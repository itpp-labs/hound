const getPatternLinkRegExp = (patternLinks) => {
    patternLinks.forEach((item) => {
        item.reg = new RegExp(item.pattern, 'g');
        item.regcopy = new RegExp(item.pattern, 'g');
    });
    const regArray = patternLinks.map((item) => item.pattern);
    return new RegExp(`(?:${ regArray.join('|') })`, 'g');
};

export const parsePatternLinks = (data) => {
    for (let repo in data) {
        if (data[repo]['pattern-links']) {
            data[repo]['pattern-link-reg'] = getPatternLinkRegExp(data[repo]['pattern-links']);
        }
    }
    return data;
};
