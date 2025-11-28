// Map Redux state to component props
export function mapStateToProps(state) {
    return {}
}

export function mapStateToPropsWebsocket(state) {
    return {
        tickerInfos: state.tickerInfos
    }
}

export function mapStateToPropsTickerBar(state) {
    return {
        tickerInfos: state.tickerInfos
    }
}

export function mapStateToPropsSubscription(state) {
    return {subscription: state.subscriptionInfo}
}

// Map Redux actions to component props
export function mapDispatchToProps(dispatch) {
    return {
        onReduxUpdate: function (key, val) {
            dispatch({ type: 'GENERAL_VALUE_UPDATE', key, value: val });
            return Promise.resolve(); // or just return void
        },
    }
}

