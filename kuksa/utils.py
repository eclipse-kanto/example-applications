# Copyright (c) 2024 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Eclipse Public License 2.0 which is available at
# https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
# which is available at https://www.apache.org/licenses/LICENSE-2.0.
#
# SPDX-License-IDentifier: EPL-2.0 OR Apache-2.0

import json

# returns a key-value pairs where the key is the VSS path and the value is the  
# value reported for this path
def process_tree(vss_tree):
    res = {}
    vss_data = json.loads(vss_tree)
    for item in vss_data:
        if item.get('value') is not None:
            res[item['path']] = item['value']
    return res


# returns a key-value pairs where the key is the VSS path and the value is the
# value reported for this path
def process_signal(vss_signal):
    res = {}
    vss_signal_json = json.loads(vss_signal)
    for item in vss_signal_json:
        entry = item['entry']
        res[entry['path']] = entry['value']['value']
    return res