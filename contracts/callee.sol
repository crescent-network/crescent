pragma solidity ^0.8.10;

//contract that is called from the caller

contract callee {
    uint public Int; //public variable initially set to 0

    function setInt(uint val) external {
        Int = val;
    }
}