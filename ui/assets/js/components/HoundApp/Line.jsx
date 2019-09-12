import React from 'react';
import { Model } from '../../helpers/Model';
import { ContentFor } from '../../utils';

export const Line = (props) => {

    const { line, rev, repo, filename, regexp } = props;
    const content = ContentFor(Model.repos[ repo ], line, regexp);

    return (
        <div className="line">
            <a href={ Model.UrlToRepo(repo, filename, line.Number, rev) }
               className="lnum"
               target="_blank"
            >
                { line.Number }
            </a>
            <span className="lval" dangerouslySetInnerHTML={ {__html:content} } />
        </div>
    );

};
