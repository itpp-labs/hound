export const FormatNumber = (t) => {
    let s = '' + (t|0);
    let b = [];
    while (s.length > 0) {
        b.unshift(s.substring(s.length - 3, s.length));
        s = s.substring(0, s.length - 3);
    }
    return b.join(',');
};

export const ParamsFromQueryString = (qs, params = {}) => {

    if (!qs) {
        return params;
    }

    qs.substring(1).split('&').forEach((v) => {
        const pair = v.split('=');
        if (pair.length != 2) {
            return;
        }

        // Handle classic '+' representation of spaces, such as is used
        // when Hound is set up in Chrome's Search Engine Manager settings
        pair[1] = pair[1].replace(/\+/g, ' ');

        params[decodeURIComponent(pair[0])] = decodeURIComponent(pair[1]);
    });

    return params;
};

export const ParamsFromUrl = (params = { q: '', i: 'nope', files: '', repos: '*' }) => ParamsFromQueryString(location.search, params);

export const ParamValueToBool = (v) => {
    v = v.toLowerCase();
    return v === 'fosho' || v === 'true' || v === '1';
};

 /**
 * Take a list of matches and turn it into a simple list of lines.
 */
export const MatchToLines = (match) => {
    const lines = [];
    const base = match.LineNumber;
    const nBefore = match.Before.length;

    match.Before.forEach((line, index) => {
        lines.push({
            Number : base - nBefore + index,
            Content: line,
            Match: false
        });
    });

    lines.push({
        Number: base,
        Content: match.Line,
        Match: true
    });

    match.After.forEach((line, index) => {
        lines.push({
            Number: base + index + 1,
            Content: line,
            Match: false
        });
    });

    return lines;
};

/**
 * Take several lists of lines each representing a matching block and merge overlapping
 * blocks together. A good example of this is when you have a match on two consecutive
 * lines. We will merge those into a singular block.
 *
 * TODO(knorton): This code is a bit skanky. I wrote it while sleepy. It can surely be
 * made simpler.
 */
export const CoalesceMatches = (matches) => {
    const blocks = matches.map(MatchToLines);
    const res = [];
    let current;
    // go through each block of lines and see if it overlaps
    // with the previous.
    for (let i = 0, n = blocks.length; i < n; i++) {
        const block = blocks[i];
        const max = current ? current[current.length - 1].Number : -1;
        // if the first line in the block is before the last line in
        // current, we'll be merging.
        if (block[0].Number <= max) {
            block.forEach((line) => {
                if (line.Number > max) {
                    current.push(line);
                } else if (current && line.Match) {
                    // we have to go back into current and make sure that matches
                    // are properly marked.
                    current[current.length - 1 - (max - line.Number)].Match = true;
                }
            });
        } else {
            if (current) {
                res.push(current);
            }
            current = block;
        }
    }

    if (current) {
        res.push(current);
    }

    return res;
};

/**
 * Use the DOM to safely htmlify some text.
 */
export const EscapeHtml = ((div) => {
    return (
        (text) => {
            div.textContent = text;
            return div.innerHTML;
        }
    );
})( document.createElement('div') );

/**
 * Produce html for a line using the regexp to highlight matches and the regexp to replace pattern-links.
 */
export const ContentFor = (repo, line, regexp) => {

    const startEm = '<em>';
    const endEm = '</em>';
    const indexes = [];

    // Store the search matches
    if (line.Match) {
        let matches;
        let len = 0;
        while ( (matches = regexp.exec(line.Content)) !== null ) {
            if (!len) { len = matches[0].length; }
            indexes.push(
                { index: matches.index, element: startEm },
                { index: matches.index + len, element: endEm }
            );
        }
        regexp.lastIndex = 0;
    }

    // Store links matches
    if (
        repo['pattern-link-reg'] &&
        repo['pattern-link-reg'].test(line.Content)
    ) {
        repo['pattern-links'].forEach((item) => {
            if ( item.reg.test(line.Content) ) {
                item.reg.lastIndex = 0;
                let matches;
                while ( (matches = item.reg.exec(line.Content)) !== null ) {
                    indexes.push(
                        {
                            index: matches.index,
                            element: `<a href="${ matches[0].replace(item.regcopy, item.link) }" target="_blank" rel="noopener noreferrer">`
                        },
                        {
                            index: matches.index + matches[0].length,
                            element: '</a>'
                        }
                    );
                }
            }
            item.reg.lastIndex = 0;
        });
    }

    if (indexes.length) {

        // Order the array
        indexes.sort((a, b) => a.index - b.index);

        let formatting = false;
        const totalIndexes = indexes.length - 1;

        return indexes.reduce((content, item, index, array) => {

            content += EscapeHtml(line.Content.slice(index ? array[index - 1].index : 0, item.index));

            if (item.element !== startEm && item.element !== endEm && formatting) {
                content += `${endEm}${item.element}${startEm}`;
            } else {
                content += item.element;
            }

            if (index === totalIndexes) {
                content += EscapeHtml(line.Content.slice(array[index].index));
            }

            if (item.element === startEm) { formatting = true; }
            if (item.element === endEm) { formatting = false; }

            return content;

        }, '');

    }

    return EscapeHtml(line.Content);
};

/**
 * Return the closest parent element
 * @param element
 * @param className
 */
export const closestElement = (element, className) => {
    while (element.className !== className) {
        element = element.parentNode;
        if (!element) {
            return null;
        }
    }
    return element;
};
