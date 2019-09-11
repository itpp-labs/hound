import React from 'react';
import { Line } from './Line';

export const Match = (props) => {

    const { block, repo, regexp, rev, filename } = props;

    const lines = block.map((line, index) => {
        return (
            <Line
                key={`line-${index}`}
                rev={ rev }
                repo={ repo }
                filename={ filename }
                regexp={ regexp }
                line={ line }
            />
        );
    });

    return (
        <div className="match">
            { lines }
        </div>
    );

};
