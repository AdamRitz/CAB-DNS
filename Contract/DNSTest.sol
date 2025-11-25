// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ITargetContract {
    function Request(string memory content, string memory IP,string memory K,string memory psk) external;
    function Assign(string memory name,bytes32 Tx) external;
}

contract SCAB {
    address SCABDNS = 0x590475c3147f58A485912A0281ee159cc79Ba5B3;

    ITargetContract target = ITargetContract(SCABDNS);

    function Request(int n ) public {
        for (int i=0;i<n;i++){
            target.Request(
            "A website providing the latest global technology news and trends.",
            "9.9.9.9"
            ,"048b406dcfd6ace48af030308a0e32f6d29d107bbdcc8843b10f4678b9672a3c23dda6f44a13adf613fb93cca77a68b2dd70e8ad0a14270f8ea044f8c2bad87bb744dc525fb873d52664065fb79ab92ec1f29ee05d146c4f53643f580fe67c93bfabd6b12962d48152264a88f4eef916c1309e664eb5b145d13dc40988af10278eeee9d749230fc339edd822929bc9c090bb1c5981289070b7b72f3e1346d0d07191524b6224b77ffa854fac9399438d3221b73775ae1a6ea84deaf796c44d0f5deb17144b64c6d0b0b5157edb71e8a06d33ed6663693720448700da625fbe7c56669c"
            ,"04ac993f4b026b44e0d98c6ab998dad768d6021bf97ac0cb70f2818419efb3867c09772ca6d3fcfd95c057049c96c8936a7eac5bd23456108d55e7d74e390655409568f6f0c9464656891c078bd6e4076b39fe32ac551f32ed8a382c065ae734a68d5336c17ac25667e36f1a26dc1b5811d64128732b97a39bfc76db16f969d6911e39a8696a5e5be161cef29f94fdf462919669340ddb223ca9325a0a15b0b2cb6678e1ea2a38f4ae338396994a"
        );

        }
        
    }
    function Assign(int n)public{
        for (int i=0;i<n;i++){
            target.Assign("GlobalNews.scab",0xd302c125ed3e967b036e8dd142ae42018a01a1be8152342e233ba6fe73c24741);
        }
    }
}
