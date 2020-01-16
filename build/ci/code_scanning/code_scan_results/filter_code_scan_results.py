from os import path
from sys import argv, stderr, exit
import json_utils as utils


def main():
    output_path, target_name = tuple(argv[1:])

    code_scan_file_path = path.join(output_path, "{}.json".format(target_name))
    code_scan_res = utils.get_deserialized_json_data(code_scan_file_path)

    if not code_scan_res:
        print("No results were found.")
    else:
        _filter_code_scan_res_gosec(code_scan_res)

    filtered_code_scan_file_path = path.join(output_path, "filtered-{}.json".format(target_name))
    utils.serialize_json_data(filtered_code_scan_file_path, code_scan_res)


def _filter_code_scan_res_gosec(code_scan_res):
    _remove_fields_dict(code_scan_res, fields_to_remove_keys=("Golang errors", "Stats"))
    _remove_inner_fields_dict(code_scan_res, key_for_inner_dicts="Issues", inner_fields_to_remove_keys=("cwe", "line"))


def _remove_fields_dict(code_scan_res, fields_to_remove_keys):
    try:
        for field_to_remove_key in fields_to_remove_keys:
            del code_scan_res[field_to_remove_key]
    except KeyError as err:
        stderr.write("Error: KeyError occurred at {}.\n".format(err))
        exit(1)


def _remove_inner_fields_dict(code_scan_res, key_for_inner_dicts, inner_fields_to_remove_keys):
    try:
        for inner_dict in code_scan_res[key_for_inner_dicts]:
            for inner_field_to_remove_key in inner_fields_to_remove_keys:
                del inner_dict[inner_field_to_remove_key]
    except KeyError as err:
        stderr.write("Error: KeyError occurred at {}.\n".format(err))
        exit(1)


if __name__ == "__main__":
    main()
