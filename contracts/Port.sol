// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

contract ProposalStore {
    struct Proposal {
        // @notice Unique id for looking up a proposal
        uint256 id;
        string title;
        string desc;
        // @notice the ordered list of target addresses for calls to be made
        address[] targets;
        uint256[] values;
        // @notice The ordered list of function signatures to be called
        string[] signatures;
        // @notice The ordered list of calldata to be passed to each call
        bytes[] calldatas;
    }

    address immutable govshuttleModAcct;

    mapping(uint256 => Proposal) private proposals;

    constructor(
        uint256 propId,
        string memory title,
        string memory desc,
        address[] memory targets,
        uint256[] memory values,
        string[] memory signatures,
        bytes[] memory calldatas
    ) {
        govshuttleModAcct = msg.sender;
        Proposal memory prop = Proposal(
            propId,
            title,
            desc,
            targets,
            values,
            signatures,
            calldatas
        );
        proposals[propId] = prop;
    }

    function AddProposal(
        uint256 propId,
        string memory title,
        string memory desc,
        address[] memory targets,
        uint256[] memory values,
        string[] memory signatures,
        bytes[] memory calldatas
    ) public {
        require(msg.sender == govshuttleModAcct); // only govshuttle account can add proposals to store
        Proposal memory newProp = Proposal(
            propId,
            title,
            desc,
            targets,
            values,
            signatures,
            calldatas
        );
        proposals[propId] = newProp;
    }

    function QueryProp(uint256 propId) public view returns (Proposal memory) {
        if (proposals[propId].id == propId) {
            return proposals[propId];
        }
        return
            Proposal(
                0,
                "",
                "",
                new address[](0),
                new uint256[](0),
                new string[](0),
                new bytes[](0)
            );
    }
}
