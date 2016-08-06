// tslint:disable

import * as React from "react";

import Component from "sourcegraph/Component";
import * as s from "sourcegraph/blob/styles/Blob.css";

class BlobToolbar extends Component<any, any> {
	reconcileState(state, props) {
		state.repo = props.repo;
		state.rev = props.rev || null;
		state.commitID = props.commitID || null;
		state.path = props.path || null;
	}

	render(): JSX.Element | null {
		return (
			<div className={s.toolbar}>
				<div className="actions">
				</div>
			</div>
		);
	}
}

(BlobToolbar as any).propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string,
	commitID: React.PropTypes.string,
	path: React.PropTypes.string,
};

export default BlobToolbar;
