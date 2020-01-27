from sys import argv
import filter_code_scan_results as filter_res


def main():
    output_path, target_name = tuple(argv[1:])

    code_scan_res = filter_res.get_code_scan_res(output_path, target_name)
    if not code_scan_res:
        print("No results were found.")
    else:
        filter_res.filter_code_scan_res_gosec(code_scan_res)
    filter_res.serialize_filtered_res(output_path, target_name, code_scan_res)


if __name__ == "__main__":
    main()
